use std::sync::Mutex;
use tauri::Manager;
use tauri_plugin_shell::process::CommandEvent;
use tauri_plugin_shell::ShellExt;

fn ensure_app_data_dir(app: &tauri::App) -> Result<std::path::PathBuf, Box<dyn std::error::Error>> {
    let dir = app.path().app_data_dir()?;
    if !dir.exists() {
        std::fs::create_dir_all(&dir)?;
    }
    Ok(dir)
}

#[derive(Default)]
struct BackendPort(String);

#[tauri::command]
fn get_backend_port(state: tauri::State<'_, Mutex<BackendPort>>) -> String {
    state.lock().unwrap().0.clone()
}

fn find_available_port(start: u16) -> Option<u16> {
    for port in start..=65535 {
        // Bind on 0.0.0.0 to match Go's :port binding behavior
        if std::net::TcpListener::bind(("0.0.0.0", port)).is_ok() {
            return Some(port);
        }
    }
    None
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_notification::init())
        .plugin(tauri_plugin_shell::init())
        .invoke_handler(tauri::generate_handler![get_backend_port])
        .setup(|app| {
            let port = match find_available_port(8000) {
                Some(p) => p,
                None => {
                    eprintln!("[tauri-setup] ERROR: no available port found");
                    return Err("no available port found".into());
                }
            };
            let port_str = port.to_string();
            println!("[tauri-setup] selected backend port: {}", port_str);

            let app_data_dir = match ensure_app_data_dir(&app) {
                Ok(dir) => dir,
                Err(e) => {
                    eprintln!("[tauri-setup] ERROR: failed to create app data dir: {}", e);
                    return Err(format!("failed to create app data dir: {}", e).into());
                }
            };
            let db_path = app_data_dir.join("stock_monitor.db");
            println!("[tauri-setup] db path: {:?}", db_path);

            let sidecar_cmd = match app.shell().sidecar("stock-monitor") {
                Ok(cmd) => cmd,
                Err(e) => {
                    eprintln!("[tauri-setup] ERROR: failed to create sidecar command: {}", e);
                    return Err(format!("failed to create sidecar command: {}", e).into());
                }
            };

            let sidecar_cmd = sidecar_cmd
                .env("PORT", &port_str)
                .env("DB_PATH", db_path.to_string_lossy().to_string())
                .env("TAURI_SIDECAR", "1");

            let (mut rx, child) = match sidecar_cmd.spawn() {
                Ok((r, c)) => (r, c),
                Err(e) => {
                    eprintln!("[tauri-setup] ERROR: failed to spawn sidecar: {}", e);
                    return Err(format!("failed to spawn sidecar: {}", e).into());
                }
            };

            println!("[tauri-setup] sidecar spawned successfully on port {}", port_str);

            // Store port and child process
            app.manage(Mutex::new(BackendPort(port_str.clone())));
            app.manage(Mutex::new(child));

            // Always log sidecar output for diagnostics
            tauri::async_runtime::spawn(async move {
                while let Some(event) = rx.recv().await {
                    match event {
                        CommandEvent::Stdout(line) => {
                            println!("[sidecar stdout] {}", String::from_utf8_lossy(&line));
                        }
                        CommandEvent::Stderr(line) => {
                            eprintln!("[sidecar stderr] {}", String::from_utf8_lossy(&line));
                        }
                        CommandEvent::Error(e) => {
                            eprintln!("[sidecar error] {}", e);
                        }
                        CommandEvent::Terminated(payload) => {
                            eprintln!(
                                "[sidecar terminated] code={:?} signal={:?}",
                                payload.code, payload.signal
                            );
                        }
                        _ => {}
                    }
                }
            });

            if cfg!(debug_assertions) {
                app.handle().plugin(
                    tauri_plugin_log::Builder::default()
                        .level(log::LevelFilter::Info)
                        .build(),
                )?;
            }
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

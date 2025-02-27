use once_cell::unsync::OnceCell;
use std::env::args;
use std::mem;
use std::rc::Rc;
use webview2::Controller;
use winapi::shared::windef::*;
use winapi::um::winuser::*;
use winit::dpi::Size;
use winit::event::{Event, WindowEvent};
use winit::event_loop::{ControlFlow, EventLoop};
use winit::platform::windows::{WindowBuilderExtWindows, WindowExtWindows};
use winit::window::{Icon, WindowBuilder};


pub async fn run(port: u16) -> Result<(), Box<dyn std::error::Error>> {
    println!("Starting webview on port {}", port);
    let event_loop = EventLoop::new();
    let image = image::load_from_memory(include_bytes!("../static/icon.ico")).expect("failed to load icon");
    let icon = Icon::from_rgba(image.to_rgba8().into_raw(), image.width(), image.height()).expect("failed to create icon");
    let window = WindowBuilder::new()
        .with_title("Snorlax | Rust Release | github.com/TrippleAWap")
        .with_inner_size(Size::Logical((1600, 900).into()))
        .with_taskbar_icon(Some(icon))
        .build(&event_loop)
        .unwrap();

    let controller: Rc<OnceCell<Controller>> = Rc::new(OnceCell::new());

    let create_result = {
        let controller_clone = controller.clone();
        let hwnd = window.hwnd() as HWND;
        webview2::Environment::builder().build(move |env| {
            env.expect("env")
                .create_controller(hwnd, move |controller| {
                    let controller = controller.expect("create host");
                    let w = controller.get_webview().expect("get_webview");
                    let _ = w.get_settings().map(|settings| {
                        let _ = settings.put_is_status_bar_enabled(true);
                        let _ = settings.put_are_default_context_menus_enabled(true);
                        let _ = settings.put_is_zoom_control_enabled(true);
                        let _ = settings.put_are_dev_tools_enabled(args().any(|arg| arg == "--devtools"));
                    });
                    w.add_script_to_execute_on_document_created(&format!("window.port = {}", port.to_string()), |err| {
                        println!("{}", err);
                        Ok(())
                    }).expect("execute_script");
                    w.add_script_to_execute_on_document_created(include_str!("../static/inject.js"),  |_| {
                        Ok(())
                    }).expect("failed to add inject.js");
                    w.navigate_to_string(include_str!("../static/index.html")).expect("failed to embed HTML");
                    unsafe {
                        let mut rect = mem::zeroed();
                        GetClientRect(hwnd, &mut rect);
                        controller.put_bounds(rect).expect("put_bounds");
                    }

                    controller_clone.set(controller).unwrap();
                    Ok(())
                })
        })
    };
    if let Err(e) = create_result {
        eprintln!(
            "Failed to create webview environment: {}. Is the new edge browser installed?",
            e
        );
    }

    event_loop.run(move |event, _, control_flow| {
        *control_flow = ControlFlow::Wait;

        match event {
            Event::WindowEvent { event, .. } => match event {
                WindowEvent::CloseRequested => {
                    if let Some(webview_host) = controller.get() {
                        webview_host.close().expect("close");
                    }
                    *control_flow = ControlFlow::Exit;
                }
                // Notify the webview when the parent window is moved.
                WindowEvent::Moved(_) => {
                    if let Some(webview_host) = controller.get() {
                        let _ = webview_host.notify_parent_window_position_changed();
                    }
                }
                // Update webview bounds when the parent window is resized.
                WindowEvent::Resized(new_size) => {
                    if let Some(webview_host) = controller.get() {
                        let r = RECT {
                            left: 0,
                            top: 0,
                            right: new_size.width as i32,
                            bottom: new_size.height as i32,
                        };
                        webview_host.put_bounds(r).expect("put_bounds");
                    }
                }
                _ => {}
            },
            Event::MainEventsCleared => {
                // Application update code.

                // Queue a RedrawRequested event.
                window.request_redraw();
            }
            Event::RedrawRequested(_) => {}
            Event::UserEvent(any) => {
                println!("{:?}", any);
            }
            _ => (),
        }
    });
}

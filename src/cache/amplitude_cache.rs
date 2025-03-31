use notify::{Config, DebouncedEvent, PollWatcher, RecursiveMode, Watcher};
use std::path::Path;
use std::sync::mpsc::channel;
use std::time::Duration;

pub fn watch_amplitude_cache() {
    let (tx, rx) = channel();

    let mut watcher = PollWatcher::new(
        tx,
        Config::default().with_poll_interval(Duration::from_millis(50)),
    )
    .unwrap();

    let path = Path::new("path/to/your/directory");

    // Start watching the directory
    watcher.watch(path, RecursiveMode::Recursive).unwrap();

    println!("Watching for changes in {:?}", path);

    // Loop to handle events
    loop {
        match rx.recv() {
            Ok(event) => match event {
                _ => println!("Received event: {:?}", event),
            },
            Err(e) => println!("Error receiving event: {:?}", e),
        }
    }
}

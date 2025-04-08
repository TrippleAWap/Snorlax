use std::collections::HashMap;
use std::env;
use notify::{Config, EventKind, PollWatcher, RecursiveMode, Watcher};
use std::sync::mpsc::channel;
use std::time::Duration;
use notify::event::{MetadataKind, ModifyKind};
use crate::cache::cache_windows_player::get_avatar_ids;
use crate::cache::scrape::process_avatars;

pub async fn watch_amplitude_cache() -> Result<(), Box<dyn std::error::Error>> {
    let (tx, rx) = channel();

    let mut watcher = PollWatcher::new(
        tx,
        Config::default().with_poll_interval(Duration::from_millis(50)),
    )?;

    let path = env::var("temp")? + "\\VRChat\\VRChat\\amplitude.cache";
    println!("Watching Amplitude Cache {:#?}", path);
    watcher.watch(path.as_ref(), RecursiveMode::NonRecursive)?;

    loop {
        match rx.recv()? {
            Ok(event) => match event.kind {
                EventKind::Modify(ModifyKind::Metadata(MetadataKind::WriteTime)) => {
                    println!("Amplitude Cache modified: {:?}", event.paths);
                    let avatar_ids = get_avatar_ids(path.as_ref(), false).await;
                    let mut avatar_map = HashMap::new();
                    println!("Found {} avatar ids", avatar_ids.len());
                    for id in avatar_ids {
                        avatar_map.insert(id.to_string(), id);
                    }
                    tokio::spawn(async {
                        process_avatars(env::var("AUTH_TOKEN").ok(), avatar_map).await.expect("Error processing data file in background");
                    });
                }
                _ => println!("Received event: {:?}", event),
            },
            Err(e) => println!("Error receiving event: {:?}", e),
        }
    }
}

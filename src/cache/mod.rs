mod scrape_cache;

use tokio::time::{sleep, Duration};

pub async fn run() -> Result<(), Box<dyn std::error::Error>> {
    // Simulate some async work
    sleep(Duration::from_secs(1)).await;
    println!("Cache module running...");
    Ok(())
}

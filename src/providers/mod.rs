mod provider;
mod prismic_provider;

use rusqlite::Connection;
use tokio::sync::Mutex;
use prismic_provider::PrismicProvider;
use crate::providers::provider::Provider;

pub async fn run() -> Result<(), Box<dyn std::error::Error>>  {
    let prismic_provider = PrismicProvider{
        database_connection: Mutex::new(Connection::open("prismic.db").unwrap()),
        table_name: "avatars".to_string(),
    };
    prismic_provider.on_start().await;
    prismic_provider.spawn_thread().await;
    Ok(())
}
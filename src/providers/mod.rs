pub mod provider;
pub mod prismic_provider;

use lazy_static::lazy_static;
use rusqlite::Connection;
use tokio::sync::Mutex;
use prismic_provider::PrismicProvider;
use crate::providers::provider::Provider;

lazy_static! {
    pub static ref PRISMIC_PROVIDER: PrismicProvider = PrismicProvider{
        database_connection: Mutex::new(Connection::open("prismic.db").unwrap()),
        table_name: "avatars".to_string(),
    };
}
#[allow(unreachable_code)]
pub async fn run() -> Result<(), Box<dyn std::error::Error>>  {
    return Ok(()); // TODO: Finish provider implementation
    PRISMIC_PROVIDER.on_start().await;
    PRISMIC_PROVIDER.spawn_thread().await;
    Ok(())
}
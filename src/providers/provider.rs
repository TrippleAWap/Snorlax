use std::fmt::Debug;
use rusqlite::{Connection};
use tokio::sync::Mutex;

#[derive(Debug)]
pub struct DatabaseEntry {
    pub avatar_id: String,

    pub avatar_name: String,
    pub avatar_description: String,

    pub author_id: String,
}

pub trait Provider: Debug {
    async fn on_start(&self) -> ();
    fn database_connection(&self) -> &Mutex<Connection>;
    async fn get_entries(&self) -> Result<Vec<DatabaseEntry>, rusqlite::Error>;
    async fn add_entries(&self, entries: Vec<DatabaseEntry>) -> Result<(), rusqlite::Error>;
    async fn spawn_thread(&self) -> (); // this should be called once per provider instance
}


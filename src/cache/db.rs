use rusqlite::{params, Connection, Result};
use log::info;
use tokio::sync::Mutex;
use lazy_static::lazy_static;

lazy_static! {
    pub static ref CONN: Mutex<Connection> = Mutex::new(Connection::open("./cache.db").unwrap());
}

pub async fn init_db(table: &str) -> Result<()> {
    let conn = CONN.lock().await;
    conn.execute(
        &format!("CREATE TABLE IF NOT EXISTS {} (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            key TEXT UNIQUE NOT NULL,
            value TEXT NOT NULL,

            insertion_timestamp TEXT DEFAULT CURRENT_TIMESTAMP
        )", table),
        params![],
    ).expect("Error creating {} table");
    let mut stmt = conn.prepare(&format!("SELECT COUNT(*) FROM {}", table)).expect("Error preparing statement");
    let count: u64 = stmt.query_row(params![], |row| row.get(0)).expect(&format!("Error getting count from {} table", table));
    info!("{} table has {} entries", table, count);

    Ok(())
}


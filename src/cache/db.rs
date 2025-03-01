use rusqlite::{params, Connection, Result};
use log::info;
use std::sync::Mutex;
use lazy_static::lazy_static;
lazy_static! {
    pub static ref CONN: Mutex<Connection> = Mutex::new(Connection::open("./cache.db").unwrap());
}

pub fn init_db(table: &str) -> Result<()> {
    let conn = CONN.lock().unwrap();
    conn.execute(
        &format!("CREATE TABLE IF NOT EXISTS {} (
            key TEXT PRIMARY KEY,
            value TEXT NOT NULL
        )", table),
        params![],
    ).expect("Error creating {} table");
    let mut stmt = conn.prepare(&format!("SELECT COUNT(*) FROM {}", table)).expect("Error preparing statement");
    let count: u64 = stmt.query_row(params![], |row| row.get(0)).expect(&format!("Error getting count from {} table", table));
    info!("{} table has {} entries", table, count);

    Ok(())
}

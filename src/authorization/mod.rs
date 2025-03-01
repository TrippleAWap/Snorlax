use lazy_static::lazy_static;
use rusqlite::{Connection, Result};
use std::sync::Mutex;

lazy_static! {
    static ref AUTH_DB: Mutex<Connection> = Mutex::new(Connection::open("./auth.db").expect("Couldn't open cache.db"));
}

pub async fn run() -> Result<(), Box<dyn std::error::Error>> {
    println!("Authorization module running...");
    Ok(())
}

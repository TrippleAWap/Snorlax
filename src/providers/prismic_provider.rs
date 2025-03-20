use std::fmt::Debug;
use rusqlite::{params, Connection, Error};
use tokio::sync::Mutex;
use crate::providers::provider::{DatabaseEntry, Provider};

#[derive(Debug)]
pub struct PrismicProvider {
    pub(crate) database_connection: Mutex<Connection>,
    pub(crate) table_name: String,
}
const PRISMIC_DATABASE_URL: &str = "https://gist.githubusercontent.com/Mwr247/ef9a06ee1d3209a558b05561f7332d8e/raw/vrcavtrdb.txt";

async fn get_prismic_database() -> Vec<DatabaseEntry> {
    let mut entries = vec![];
    let response = reqwest::get(PRISMIC_DATABASE_URL).await.unwrap().text().await.unwrap();

    let lines = response.split('\n');
    for line in lines {
        let line = line.chars().rev().collect::<String>();
        let parts: Vec<&str> = line.split('\t').collect();
        if parts.len() == 4 {
            let avatar_id = decode_avatar_id(parts[0]);
            let avatar_name = parts[1].to_string();
            let author_id = parts[2].to_string();
            let avatar_description = parts[3].to_string();
            entries.push(DatabaseEntry {
                avatar_id,
                avatar_name,
                avatar_description,
                author_id,
            });
        }
    }

    entries
}

fn decode_avatar_id(crypt: &str) -> String {
    const CIPHER: &str = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+";

    let mut decrypt: Vec<char> = vec![' '; 33];
    let new_format = (CIPHER.chars().position(|c| c == crypt.chars().nth(21).unwrap()).unwrap() >> 2) & 2;

    for i in 0..11 {
        let idx = i * 3;
        let first = CIPHER.chars().position(|c| c == crypt.chars().nth(i * 2).unwrap()).unwrap();
        let third = CIPHER.chars().position(|c| c == crypt.chars().nth(i * 2 + 1).unwrap()).unwrap();

        decrypt[idx] = CIPHER.chars().nth((first >> new_format) & 15).unwrap();
        let second = if new_format == 0 {
            (first >> 2) & 12
        } else {
            (first & 3) << 2
        };

        decrypt[idx + 1] = CIPHER.chars().nth(second | ((third >> 4) & 3)).unwrap();
        decrypt[idx + 2] = CIPHER.chars().nth(third & 15).unwrap();
    }

    decrypt.pop();

    decrypt.insert(8, '-');
    decrypt.insert(13, '-');
    decrypt.insert(18, '-');
    decrypt.insert(23, '-');

    let decrypted_str: String = decrypt.into_iter().collect();

    format!("avtr_{}", decrypted_str)
}

impl Provider for PrismicProvider {
    async fn on_start(&self) -> () {
        let conn = self.database_connection().lock().await;
        conn.execute(
            &format!("CREATE TABLE IF NOT EXISTS {} (
            id INTEGER PRIMARY KEY AUTOINCREMENT,

            avatar_id TEXT UNIQUE NOT NULL,

            avatar_name TEXT NOT NULL,
            avatar_description TEXT NOT NULL,

            author_id TEXT NOT NULL,
        )", self.table_name),
            params![],
        )
       .unwrap();
    }
    fn database_connection(&self) -> &Mutex<Connection> {
        &self.database_connection
    }

    async fn get_entries(&self) -> Result<Vec<DatabaseEntry>, Error> {
        let conn = self.database_connection().lock().await;
        let mut stmt = conn.prepare(&format!("SELECT * FROM {}", self.table_name))?;
        let entries = stmt.query_map(params![], |row| {
            Ok(DatabaseEntry {
                avatar_id: row.get(0)?,
                avatar_name: row.get(1)?,
                avatar_description: row.get(2)?,
                author_id: row.get(3)?,
            })
        })?;
        let mut result = Vec::new();
        for entry in entries {
            result.push(entry?);
        }
        Ok(result)
    }

    async fn add_entries(&self, entries: Vec<DatabaseEntry>) -> Result<(), Error> {
        let conn = self.database_connection().lock().await;
        let mut stmt = conn.prepare(&format!("INSERT INTO {} (avatar_id, avatar_name, avatar_description, author_id) VALUES (?,?,?,?)", self.table_name))?;
        for entry in entries {
            stmt.execute(
                params![
                    entry.avatar_id,
                    entry.avatar_name,
                    entry.avatar_description,
                    entry.author_id,
                ],
            )?;
        }
        Ok(())
    }

    async fn spawn_thread(&self) -> () {
        let db = get_prismic_database().await;
        println!("Adding {} entries to the database...", db.len());
        self.add_entries(db).await.unwrap();
        println!("Database updated successfully!");
    }
}


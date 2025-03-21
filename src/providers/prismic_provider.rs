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
        let parts: Vec<&str> = line.split('\t').rev().collect();
        // println!("{:?}", parts.len());
        if parts.len() == 4 {
            let avatar_id = decode_avatar_id(parts[0].chars().rev().collect::<String>().as_str());
            if avatar_id.is_err() {
                // println!("Invalid avatar id {} - {}", parts[0], avatar_id.err().unwrap());
                continue;
            }
            let avatar_name = parts[1].to_string();
            let author_id = parts[2].to_string();
            let avatar_description = parts[3].to_string();
            let entry = DatabaseEntry {
                avatar_id: avatar_id.unwrap(),
                avatar_name,
                avatar_description,
                author_id,
            };
            println!("{:?}", entry);
            // avtr_a1c59c2f-b11e-48ad-9ba7-d7e0fede57d3
            // avtr_0009c2a9-265d-497a-7ba8-6912bb83ee5c
            entries.push(entry);
        }
    }

    entries
}

const CIPHER: &str = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+=";

fn decode_avatar_id(crypt: &str) -> Result<String, Box<dyn std::error::Error>> {
    if crypt.len() != 24 {
        return Err("Invalid input length".into());
    }

    let mut decrypt: Vec<u8> = vec![0; 33];
    let new_format = (CIPHER.find(crypt.chars().nth(21).unwrap()).unwrap() >> 2) & 2;

    for i in 0..11 {
        let a = crypt.chars().nth(i * 2).unwrap();
        let b = crypt.chars().nth(i * 2 + 1).unwrap();

        let a_index = CIPHER.find(a).ok_or("Invalid character in input")?;
        let b_index = CIPHER.find(b).ok_or("Invalid character in input")?;

        let idx = i * 3;

        let val1 = ((a_index >> new_format) & 15) as u8;
        let val2 = if new_format == 0 {
            (((a_index >> 2) & 12) | ((b_index >> 4) & 3)) as u8
        } else {
            ((((a_index & 3) << 2) | ((b_index >> 4) & 3))) as u8
        };
        let val3 = (b_index & 15) as u8;

        decrypt[idx] = CIPHER.as_bytes()[val1];
        decrypt[idx + 1] = CIPHER.as_bytes()[val2];
        decrypt[idx + 2] = CIPHER.as_bytes()[val3];
    }

    decrypt.pop();

    let mut uuid = String::new();
    for (i, &c) in decrypt.iter().enumerate() {
        if i == 8 || i == 13 || i == 18 || i == 23 {
            uuid.push('-');
        }
        uuid.push(c as char);
    }

    Ok(format!("avtr_{}", uuid))
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

            author_id TEXT NOT NULL
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


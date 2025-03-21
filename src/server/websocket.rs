use std::collections::{HashMap, HashSet};
use futures_util::{SinkExt, StreamExt};
use crate::cache::db::CONN;
use crate::server::{query_avatar_cache, CARD_HTML};
use std::sync::Mutex;
use once_cell::sync::Lazy;
use crate::server::favorites::FAVORITE_FILTER;
pub static FILTER_STRING: Lazy<Mutex<String>> = Lazy::new(|| Mutex::new(String::new()));

pub async fn process_ws(ws: warp::ws::WebSocket)  {
    println!("New connection");
    let (mut write, mut read) = ws.split();
    while let Some(msg) = read.next().await {
        let msg = msg.expect("Error receiving message");
        if msg.is_text() {
            let json_text = msg.to_str().expect("Error converting message to text");
            let json: serde_json::Value = serde_json::from_str(&json_text).expect("Error parsing JSON");
            let response: serde_json::Value;
            match json.get("event").expect("Error getting event").as_str().expect("Error") {
                "fetch_avatars" => {
                    let data = json.get("data").expect("Error").as_object().expect("Error");
                    let start = data.get("start").expect("Error").as_u64().expect("Error");
                    let end = data.get("end").expect("Error").as_u64().expect("Error");

                    let conn = CONN.lock().await;
                    let filter_string = FILTER_STRING.lock().unwrap().clone();
                    let favorites = FAVORITE_FILTER.lock().await;

                    let mut avatars = query_avatar_cache(&conn, if filter_string.is_empty() { None } else { Some(filter_string) }, *favorites).expect("Error querying avatar cache");
                    // splice the avatars into the requested range
                    let total_count = avatars.len() as u64;
                    avatars = avatars.iter().skip(start as usize).take(end as usize).cloned().collect();
                    let package_to_display_name: HashMap<String, String> = {
                        let mut map = HashMap::new();

                        map.insert("standalonewindows".to_string(), "Desktop".to_string());
                        map.insert("android".to_string(), "Quest".to_string());
                        map.insert("ios".to_string(), "IOS".to_string());

                        map
                    };

                    let mut avatar_data = avatars.iter().map(|(_, avatar_id, avatar_value)| {
                        // Parse the JSON string from the cached avatar.
                        let avatar: vrchatapi::models::Avatar = serde_json::from_str(avatar_value)
                            .expect("Error parsing avatar JSON");

                        let mut card_html = CARD_HTML.to_string()
                            .replace("{{thumbnail_url}}", &uri_encode::encode_uri_component(&avatar.thumbnail_image_url))
                            .replace("{{author_name}}", &avatar.author_name)
                            .replace("{{author_id}}", &avatar.author_id)
                            .replace("{{avatar_name}}", &avatar.name)
                            .replace("{{avatar_id}}", avatar_id);

                        let unique_platforms: HashSet<String> = avatar
                            .unity_packages
                            .iter()
                            .filter(|package| {
                                let status = package.scan_status.clone().unwrap_or_default();
                                status == "passed"
                            })
                            .map(|package| package.platform.clone())
                            .collect();

                        let mut platform_elements = unique_platforms
                            .iter()
                            .map(|platform| {
                                let display_name = package_to_display_name.get(platform).unwrap_or(platform);
                                format!(
                                    "<li class=\"inline-flex items-center px-2 py-1 text-xs font-semibold transition-colors rounded-full bg-white/5 hover:bg-white/10 text-white\">#{}</li>",
                                    display_name
                                )
                            })
                            .collect::<Vec<String>>();
                        // this seems inefficient, but I don't know any better way.
                        platform_elements.sort_by_key(|s| s.len());
                        platform_elements.reverse();
                        let platform_elements = platform_elements.join("");
                        card_html = card_html.replace("{{platform_elements}}", &platform_elements);
                        card_html
                    }).collect::<Vec<String>>();
                    avatar_data.reverse();
                    response = serde_json::json!({
                        "event": "avatars",
                        "data": serde_json::json!({
                            "avatars": avatar_data,
                            "total_count": total_count
                        })
                    });
                }
                _ => {
                    println!("Unknown event: {}", json.get("event").expect("Error getting event"));
                    response = serde_json::json!({
                        "event": "error",
                        "data": "Unknown event"
                    });
                }
            }
            let response_text = serde_json::to_string(&response).expect("Error converting JSON to text");
            let response_msg = warp::ws::Message::text(response_text);
            write.send(response_msg.clone()).await.expect("Error sending message");
        }
    }
}

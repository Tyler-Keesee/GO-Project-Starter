use log::{info, LevelFilter};
use env_logger;

const VERSION: &str = "0.1.0";

fn main() {
    env_logger::builder().filter_level(LevelFilter::Info).init();
    info!("Starting Go Test Project v{}", VERSION);
    println!("Go Test Project is ready to rock!");
}

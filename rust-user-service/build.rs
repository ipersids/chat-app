use std::fs::create_dir_all;
use std::path::Path;
use tonic_prost_build;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let out_dir = std::env::var("OUT_DIR").unwrap();
    let out_path = Path::new(&out_dir).join("proto-generated");

    create_dir_all(&out_path).unwrap();

    tonic_prost_build::configure()
        .build_server(false)
        .out_dir(&out_path)
        .compile_protos(&["../proto/user.proto"], &["../proto"])?;

    Ok(())
}

use rust_user_service::auth::AuthServiceServer;
use rust_user_service::server::AuthServiceImpl;
use std::net::SocketAddr;
use tonic::transport::Server;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let addr: SocketAddr = "127.0.0.1:50051".parse().unwrap();
    let auth_service = AuthServiceImpl::new();

    Server::builder()
        .add_service(AuthServiceServer::new(auth_service))
        .serve(addr)
        .await?;

    Ok(())
}

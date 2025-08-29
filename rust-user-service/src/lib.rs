/// Generated protobuf modules (internal)
mod proto {
    include!(concat!(env!("OUT_DIR"), "/proto-generated/user.v1.rs"));
}

/// Authentication service types and client
pub mod auth {
    pub use crate::proto::auth_service_client::AuthServiceClient;
    pub use crate::proto::{CreateRequest, CreateResponse, LoginRequest, LoginResponse, User};
}

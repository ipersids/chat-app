/// Generated protobuf modules (internal)
mod proto {
    include!(concat!(env!("OUT_DIR"), "/proto-generated/user.v1.rs"));
}

/// Authentication service types and server
pub mod auth {
    pub use crate::proto::auth_service_server::{
        AuthService as AuthServiceTrait, AuthServiceServer,
    };
    pub use crate::proto::{CreateRequest, CreateResponse, LoginRequest, LoginResponse, User};
}

pub mod server;
pub mod storage;

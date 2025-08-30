use crate::auth::{AuthServiceTrait, CreateRequest, CreateResponse, LoginRequest, LoginResponse};
use crate::storage::UserStorage;
use tonic::{Request, Response, Status};

#[derive(Debug)]
pub struct AuthServiceImpl {
    db_users: UserStorage,
}

impl AuthServiceImpl {
    pub fn new() -> Self {
        Self {
            db_users: UserStorage::new(),
        }
    }
}

#[tonic::async_trait]
impl AuthServiceTrait for AuthServiceImpl {
    async fn create(
        &self,
        request: Request<CreateRequest>,
    ) -> Result<Response<CreateResponse>, Status> {
        println!("Got a create request: {:?}", request);

        let request = request.into_inner();

        if request.login.is_empty() {
            return Err(Status::invalid_argument("Invalid login"));
        }

        if request.password.is_empty() {
            return Err(Status::invalid_argument("Invalid password"));
        }

        match self.db_users.create_user(&request.login, &request.password) {
            Ok(user) => {
                let response = CreateResponse {
                    success: true,
                    user: Some(user),
                    error: String::new(),
                };
                Ok(Response::new(response))
            }
            Err(error_msg) => {
                let response = CreateResponse {
                    success: false,
                    user: None,
                    error: error_msg.to_owned(),
                };
                Ok(Response::new(response))
            }
        }
    }

    async fn login(
        &self,
        request: Request<LoginRequest>,
    ) -> Result<Response<LoginResponse>, Status> {
        println!("Got a login request: {:?}", request);

        let request = request.into_inner();

        if request.login.is_empty() {
            return Err(Status::invalid_argument("Invalid login"));
        }

        if request.password.is_empty() {
            return Err(Status::invalid_argument("Invalid password"));
        }

        match self
            .db_users
            .authenticate_user(&request.login, &request.password)
        {
            Ok(user) => {
                let response = LoginResponse {
                    success: true,
                    user: Some(user),
                    error: String::new(),
                };
                Ok(Response::new(response))
            }
            Err(error_msg) => {
                let response = LoginResponse {
                    success: false,
                    user: None,
                    error: error_msg.to_owned(),
                };
                Ok(Response::new(response))
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn auth_service_impl_test() {
        let db = AuthServiceImpl::new();
        let req = Request::new(CreateRequest {
            login: "Cat".to_string(),
            password: "1234aasff-".to_string(),
        });
        let res = db.create(req).await;
        assert!(res.is_ok());
        let res = res.unwrap().into_inner();
        assert!(res.success);
        assert_eq!(res.user.unwrap().login, "Cat");
        let req = Request::new(CreateRequest {
            login: "Cat".to_string(),
            password: "1234aasff-".to_string(),
        });
        let res = db.create(req).await;
        assert!(res.is_ok());
        let res = res.unwrap().into_inner();
        assert!(!res.success);

        let req = Request::new(LoginRequest {
            login: "Cat".to_string(),
            password: "1234aasff-".to_string(),
        });
        let res = db.login(req).await;
        assert!(res.is_ok());
        let res = res.unwrap().into_inner();
        assert!(res.success);
        let req = Request::new(LoginRequest {
            login: "Cat".to_string(),
            password: "175chkbgfxdfzxfgvbhjff-".to_string(),
        });
        let res = db.login(req).await;
        assert!(res.is_ok());
        let res = res.unwrap().into_inner();
        assert!(!res.success);
    }
}

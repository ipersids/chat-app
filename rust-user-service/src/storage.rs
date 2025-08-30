use crate::auth::User;
use bcrypt::{DEFAULT_COST, hash, verify};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use uuid::Uuid;

#[derive(Debug, Clone)]
pub struct UserStorage {
    // HashMap: login -> (password_hash, User)
    db_users: Arc<Mutex<HashMap<String, (String, User)>>>,
}

impl UserStorage {
    pub fn new() -> Self {
        Self {
            db_users: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    /// Create a new user
    pub fn create_user(&self, login: &str, password: &str) -> Result<User, String> {
        println!("Create user '{}'", login);

        let mut db_users = self
            .db_users
            .lock()
            .map_err(|_| "Failed to get database lock".to_string())?;

        if db_users.contains_key(login) {
            println!("User '{}' already exists", login);
            return Err("User already exists".to_string());
        }

        let password_hash =
            hash(password, DEFAULT_COST).map_err(|e| format!("Invalid password: {}", e))?;

        let new_user = User {
            uuid: Uuid::new_v4().to_string(),
            login: login.to_string(),
        };

        db_users.insert(login.to_string(), (password_hash, new_user.clone()));

        println!("User '{}' created.", login);
        Ok(new_user)
    }

    /// Authenticate user login
    pub fn authenticate_user(&self, login: &str, password: &str) -> Result<User, String> {
        println!("Authenticate user '{}'", login);

        let db_users = self
            .db_users
            .lock()
            .map_err(|_| "Failed to get database lock".to_string())?;

        match db_users.get(login) {
            Some((password_hash, user)) => match verify(password, password_hash) {
                Ok(is_valid) => {
                    if is_valid {
                        println!("Authentication successful for '{}'", login);
                        Ok(user.clone())
                    } else {
                        println!("Invalid password for '{}'", login);
                        Err("Invalid credentials".to_string())
                    }
                }
                Err(_) => {
                    println!("Invalid password for '{}'", login);
                    Err("Authentication system error".to_string())
                }
            },
            None => {
                println!("User '{}' not found", login);
                Err("User not found".to_string())
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn storage_test() {
        let db = UserStorage::new();

        // create user
        let res = db.create_user("Cat", "123asdf!");
        assert!(res.is_ok());
        assert_eq!(res.unwrap().login, "Cat");

        // try to duplicate user creation
        let res = db.create_user("Cat", "dfkfjhrjfhY7");
        assert!(res.is_err());
        assert_eq!(res.unwrap_err(), "User already exists");

        // successful authentication
        let res = db.authenticate_user("Cat", "123asdf!");
        assert!(res.is_ok());
        assert_eq!(res.unwrap().login, "Cat");

        // authentication with wrong password
        let res = db.authenticate_user("Cat", "wrongpassword");
        assert!(res.is_err());
        assert_eq!(res.unwrap_err(), "Invalid credentials");

        // authentication for non-existent user
        let res = db.authenticate_user("Dog", "123asdf!");
        assert!(res.is_err());
        assert_eq!(res.unwrap_err(), "User not found");
    }

    #[test]
    fn password_hashing_test() {
        let db = UserStorage::new();

        // create user with password
        let res = db.create_user("TestUser", "mypassword123");
        assert!(res.is_ok());

        // authenticate with correct password
        let res = db.authenticate_user("TestUser", "mypassword123");
        assert!(res.is_ok());

        // authenticate with incorrect password
        let res = db.authenticate_user("TestUser", "wrongpassword");
        assert!(res.is_err());
    }
}

API Endpoints
    1. Account
        POST /accounts/register - User registration
        POST /accounts/login - User authentication
        POST /accounts/logout - User logout
        POST /accounts/ - Create new account (protected)
        GET /accounts/{id} - Get account by ID (protected)
        POST /accounts/{id} - Update account (protected)
        DELETE /accounts/{id} - Delete account (protected)
        GET /accounts/email/{email} - Find by email (protected)
Postman
    Base URL: http://localhost:8888

postgres
    psql -U postgres
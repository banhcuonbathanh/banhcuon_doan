import { useAuth } from "../../../../auth/useauth";

function LoginComponent() {
  const { login, loading, error, user } = useAuth();

  const handleLogin = async (email: string, password: string) => {
    await login({ email, password });
  };

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;
  if (user) return <div>Welcome, {user.name}!</div>;

  return <div>account page</div>;
}

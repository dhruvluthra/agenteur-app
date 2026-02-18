import { Link } from "react-router-dom"
import { Button } from "@/components/ui/button"

function Login() {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-6">
      <h1 className="text-3xl font-bold tracking-tight">Log in</h1>
      <p className="text-muted-foreground">Login form coming soon.</p>
      <Button variant="link" asChild>
        <Link to="/signup">Don't have an account? Sign up</Link>
      </Button>
    </div>
  )
}

export default Login

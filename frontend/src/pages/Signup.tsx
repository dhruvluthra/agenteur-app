import { Link } from "react-router-dom"
import { Button } from "@/components/ui/button"

function Signup() {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-6">
      <h1 className="text-3xl font-bold tracking-tight">Sign up</h1>
      <p className="text-muted-foreground">Signup form coming soon.</p>
      <Button variant="link" asChild>
        <Link to="/login">Already have an account? Log in</Link>
      </Button>
    </div>
  )
}

export default Signup

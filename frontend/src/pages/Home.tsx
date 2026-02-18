import { Link } from "react-router-dom"
import { Button } from "@/components/ui/button"

function Home() {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-6">
      <h1 className="text-4xl font-bold tracking-tight">Agenteur</h1>
      <p className="text-muted-foreground">Deploy AI agents at scale.</p>
      <div className="flex gap-4">
        <Button asChild>
          <Link to="/login">Log in</Link>
        </Button>
        <Button variant="outline" asChild>
          <Link to="/signup">Sign up</Link>
        </Button>
      </div>
    </div>
  )
}

export default Home

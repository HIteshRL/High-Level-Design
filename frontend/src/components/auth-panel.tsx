import { useMemo, useState } from "react"
import { Loader2, Sparkles } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

interface AuthPanelProps {
  onLogin: (username: string, password: string) => Promise<void>
  onRegister: (username: string, email: string, password: string) => Promise<void>
  busy: boolean
}

export function AuthPanel({ onLogin, onRegister, busy }: AuthPanelProps) {
  const [loginUsername, setLoginUsername] = useState("")
  const [loginPassword, setLoginPassword] = useState("")

  const [registerUsername, setRegisterUsername] = useState("")
  const [registerEmail, setRegisterEmail] = useState("")
  const [registerPassword, setRegisterPassword] = useState("")

  const canLogin = useMemo(() => loginUsername.trim().length > 0 && loginPassword.length >= 8, [loginUsername, loginPassword])
  const canRegister = useMemo(
    () => registerUsername.trim().length >= 3 && registerEmail.includes("@") && registerPassword.length >= 8,
    [registerUsername, registerEmail, registerPassword],
  )

  return (
    <div className="mx-auto flex min-h-screen w-full max-w-5xl items-center justify-center p-6">
      <Card className="w-full max-w-md border-border/70 shadow-lg">
        <CardHeader className="space-y-3">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Sparkles className="h-4 w-4" />
            Roognis AI Workspace
          </div>
          <CardTitle className="text-2xl">Welcome back</CardTitle>
          <CardDescription>Sign in or create an account to access the inference workspace.</CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="login">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="login">Login</TabsTrigger>
              <TabsTrigger value="register">Register</TabsTrigger>
            </TabsList>

            <TabsContent value="login" className="space-y-3 pt-3">
              <form
                className="space-y-3"
                onSubmit={(event) => {
                  event.preventDefault()
                  if (!canLogin || busy) return
                  void onLogin(loginUsername, loginPassword)
                }}
              >
                <div className="space-y-2">
                  <Label htmlFor="login-username">Username</Label>
                  <Input id="login-username" placeholder="Username" value={loginUsername} onChange={(event) => setLoginUsername(event.target.value)} />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="login-password">Password</Label>
                  <Input
                    id="login-password"
                    type="password"
                    placeholder="Password"
                    value={loginPassword}
                    onChange={(event) => setLoginPassword(event.target.value)}
                  />
                </div>
                <Button className="w-full" type="submit" disabled={!canLogin || busy}>
                  {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : "Sign in"}
                </Button>
              </form>
            </TabsContent>

            <TabsContent value="register" className="space-y-3 pt-3">
              <form
                className="space-y-3"
                onSubmit={(event) => {
                  event.preventDefault()
                  if (!canRegister || busy) return
                  void onRegister(registerUsername, registerEmail, registerPassword)
                }}
              >
                <div className="space-y-2">
                  <Label htmlFor="register-username">Username</Label>
                  <Input
                    id="register-username"
                    placeholder="Username"
                    value={registerUsername}
                    onChange={(event) => setRegisterUsername(event.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="register-email">Email</Label>
                  <Input id="register-email" type="email" placeholder="Email" value={registerEmail} onChange={(event) => setRegisterEmail(event.target.value)} />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="register-password">Password</Label>
                  <Input
                    id="register-password"
                    type="password"
                    placeholder="Password (min 8 chars)"
                    value={registerPassword}
                    onChange={(event) => setRegisterPassword(event.target.value)}
                  />
                </div>
                <Button className="w-full" type="submit" disabled={!canRegister || busy}>
                  {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : "Create account"}
                </Button>
              </form>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  )
}

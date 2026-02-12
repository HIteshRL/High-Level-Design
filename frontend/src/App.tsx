import { useEffect, useMemo, useState } from "react"
import { Toaster, toast } from "sonner"

import { AuthPanel } from "@/components/auth-panel"
import { ChatPanel } from "@/components/chat-panel"
import { ApiError, complete, listConversations, login, me, register } from "@/lib/api"
import type { Conversation, User } from "@/lib/types"

type ChatMessage = { role: "user" | "assistant"; content: string }

const TOKEN_KEY = "roognis.frontend.token"

function App() {
  const [token, setToken] = useState<string | null>(() => sessionStorage.getItem(TOKEN_KEY) ?? localStorage.getItem(TOKEN_KEY))
  const [user, setUser] = useState<User | null>(null)
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [activeConversationId, setActiveConversationId] = useState<string | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [busy, setBusy] = useState(false)

  const isAuthenticated = useMemo(() => Boolean(token && user), [token, user])

  useEffect(() => {
    if (!token) return
    const authToken = token

    let cancelled = false

    async function bootstrap() {
      try {
        const [currentUser, conversationList] = await Promise.all([me(authToken), listConversations(authToken)])
        if (cancelled) return
        setUser(currentUser)
        setConversations(conversationList)
      } catch (error) {
        if (cancelled) return
        handleError(error)
        hardLogout()
      }
    }

    void bootstrap()

    return () => {
      cancelled = true
    }
  }, [token])

  function hardLogout() {
    sessionStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(TOKEN_KEY)
    setToken(null)
    setUser(null)
    setConversations([])
    setActiveConversationId(null)
    setMessages([])
  }

  function persistToken(nextToken: string) {
    sessionStorage.setItem(TOKEN_KEY, nextToken)
    setToken(nextToken)
  }

  async function handleLogin(username: string, password: string) {
    setBusy(true)
    try {
      const auth = await login(username, password)
      persistToken(auth.access_token)
      toast.success("Signed in")
    } catch (error) {
      handleError(error)
    } finally {
      setBusy(false)
    }
  }

  async function handleRegister(username: string, email: string, password: string) {
    setBusy(true)
    try {
      const auth = await register(username, email, password)
      persistToken(auth.access_token)
      toast.success("Account created")
    } catch (error) {
      handleError(error)
    } finally {
      setBusy(false)
    }
  }

  async function refreshConversations() {
    if (!token) return
    try {
      const conversationList = await listConversations(token)
      setConversations(conversationList)
    } catch (error) {
      handleError(error)
    }
  }

  async function handleSend(prompt: string, conversationId: string | null) {
    if (!token) {
      throw new Error("Missing authentication token")
    }

    setBusy(true)
    setMessages((previous) => [...previous, { role: "user", content: prompt }])

    try {
      const response = await complete(token, {
        prompt,
        conversation_id: conversationId ?? undefined,
      })

      setMessages((previous) => [...previous, { role: "assistant", content: response.content }])
      setActiveConversationId(response.conversation_id)
      await refreshConversations()

      if (response.cached) {
        toast.message("Served from cache")
      }

      return response
    } catch (error) {
      setMessages((previous) => previous.slice(0, -1))
      handleError(error)
      throw error
    } finally {
      setBusy(false)
    }
  }

  function handleError(error: unknown) {
    if (error instanceof ApiError) {
      toast.error(error.message)
      return
    }

    if (error instanceof Error) {
      toast.error(error.message)
      return
    }

    toast.error("An unexpected error occurred")
  }

  return (
    <>
      {isAuthenticated && user ? (
        <ChatPanel
          user={user}
          conversations={conversations}
          activeConversationId={activeConversationId}
          messages={messages}
          busy={busy}
          onSelectConversation={(id) => {
            setActiveConversationId(id)
            setMessages([])
          }}
          onSend={handleSend}
          onLogout={hardLogout}
        />
      ) : (
        <AuthPanel onLogin={handleLogin} onRegister={handleRegister} busy={busy} />
      )}
      <Toaster position="top-right" richColors />
    </>
  )
}

export default App

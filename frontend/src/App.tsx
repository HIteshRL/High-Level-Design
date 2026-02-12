import { useEffect, useMemo, useState } from "react"
import { Toaster, toast } from "sonner"

import { AuthPanel } from "@/components/auth-panel"
import { ChatPanel } from "@/components/chat-panel"
import { ApiError, complete, listConversationMessages, listConversations, login, me, register } from "@/lib/api"
import type { Conversation, ConversationMessage, User } from "@/lib/types"

type ChatMessage = {
  id: string
  role: "user" | "assistant"
  content: string
  createdAt: string
  model?: string
  latencyMs?: number
  cached?: boolean
}

const TOKEN_KEY = "roognis.frontend.token"
const THEME_KEY = "roognis.frontend.theme"

type ThemeMode = "light" | "dark"

interface InferenceSettings {
  model: string
  temperature: number
  maxTokens: number
}

function App() {
  const [token, setToken] = useState<string | null>(() => sessionStorage.getItem(TOKEN_KEY) ?? localStorage.getItem(TOKEN_KEY))
  const [user, setUser] = useState<User | null>(null)
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [activeConversationId, setActiveConversationId] = useState<string | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [busy, setBusy] = useState(false)
  const [historyBusy, setHistoryBusy] = useState(false)
  const [theme, setTheme] = useState<ThemeMode>(() => {
    const stored = localStorage.getItem(THEME_KEY)
    return stored === "light" || stored === "dark" ? stored : "dark"
  })
  const [settings, setSettings] = useState<InferenceSettings>({
    model: "qwen2.5:0.5b",
    temperature: 0.7,
    maxTokens: 1024,
  })

  const isAuthenticated = useMemo(() => Boolean(token && user), [token, user])

  function normalizeConversationMessages(persisted: ConversationMessage[]): ChatMessage[] {
    return persisted
      .filter((message): message is ConversationMessage & { role: "user" | "assistant" } => message.role === "user" || message.role === "assistant")
      .map((message) => ({
        id: message.id,
        role: message.role,
        content: message.content,
        createdAt: message.created_at,
        model: message.model_used,
        latencyMs: message.latency_ms,
        cached: false,
      }))
  }

  useEffect(() => {
    const root = document.documentElement
    if (theme === "dark") {
      root.classList.add("dark")
    } else {
      root.classList.remove("dark")
    }
    localStorage.setItem(THEME_KEY, theme)
  }, [theme])

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

        if (conversationList.length > 0) {
          const nextConversationID = conversationList[0].id
          setActiveConversationId(nextConversationID)
          const persisted = await listConversationMessages(authToken, nextConversationID)
          if (cancelled) return
          setMessages(normalizeConversationMessages(persisted))
        }
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

  async function loadConversationMessages(conversationId: string) {
    if (!token) return
    setHistoryBusy(true)
    try {
      const persisted = await listConversationMessages(token, conversationId)
      setMessages(normalizeConversationMessages(persisted))
    } catch (error) {
      handleError(error)
      setMessages([])
    } finally {
      setHistoryBusy(false)
    }
  }

  async function handleSend(prompt: string, conversationId: string | null) {
    if (!token) {
      throw new Error("Missing authentication token")
    }

    setBusy(true)
    const userMessageId = crypto.randomUUID()
    setMessages((previous) => [
      ...previous,
      { id: userMessageId, role: "user", content: prompt, createdAt: new Date().toISOString() },
    ])

    try {
      const response = await complete(token, {
        prompt,
        conversation_id: conversationId ?? undefined,
        model: settings.model,
        temperature: settings.temperature,
        max_tokens: settings.maxTokens,
      })

      setMessages((previous) => [
        ...previous,
        {
          id: response.id,
          role: "assistant",
          content: response.content,
          createdAt: new Date().toISOString(),
          model: response.model,
          latencyMs: response.latency_ms,
          cached: response.cached,
        },
      ])
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
          historyBusy={historyBusy}
          theme={theme}
          settings={settings}
          onSelectConversation={(id) => {
            if (!id) {
              setActiveConversationId(null)
              setMessages([])
              return
            }
            setActiveConversationId(id)
            void loadConversationMessages(id)
          }}
          onThemeChange={setTheme}
          onSettingsChange={setSettings}
          onSend={handleSend}
          onLogout={hardLogout}
        />
      ) : (
        <AuthPanel onLogin={handleLogin} onRegister={handleRegister} busy={busy} theme={theme} onThemeChange={setTheme} />
      )}
      <Toaster position="top-right" richColors />
    </>
  )
}

export default App

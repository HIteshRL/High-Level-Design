import { useEffect, useMemo, useState } from "react"
import { Bot, Copy, Gauge, Loader2, LogOut, Moon, PanelLeftClose, PanelLeftOpen, Send, Sparkles, Sun, User as UserIcon } from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { MarkdownContent } from "@/components/markdown-content"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Separator } from "@/components/ui/separator"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"
import type { Conversation, InferenceResponse, User } from "@/lib/types"
import { cn } from "@/lib/utils"

interface ChatPanelProps {
  user: User
  conversations: Conversation[]
  activeConversationId: string | null
  messages: Array<{
    id: string
    role: "user" | "assistant"
    content: string
    createdAt: string
    model?: string
    latencyMs?: number
    cached?: boolean
  }>
  busy: boolean
  historyBusy: boolean
  theme: "light" | "dark"
  settings: {
    model: string
    temperature: number
    maxTokens: number
  }
  onSelectConversation: (id: string | null) => void
  onThemeChange: (theme: "light" | "dark") => void
  onSettingsChange: (settings: { model: string; temperature: number; maxTokens: number }) => void
  onSend: (prompt: string, conversationId: string | null) => Promise<InferenceResponse>
  onLogout: () => void
}

export function ChatPanel({
  user,
  conversations,
  activeConversationId,
  messages,
  busy,
  historyBusy,
  theme,
  settings,
  onSelectConversation,
  onThemeChange,
  onSettingsChange,
  onSend,
  onLogout,
}: ChatPanelProps) {
  const sidebarWidthKey = "roognis.frontend.sidebar.width"
  const sidebarCollapsedKey = "roognis.frontend.sidebar.collapsed"
  const minSidebarWidth = 240
  const maxSidebarWidth = 420
  const defaultSidebarWidth = 280
  const promptLimit = 32000
  const [prompt, setPrompt] = useState("")
  const [sidebarWidth, setSidebarWidth] = useState(() => {
    const raw = localStorage.getItem(sidebarWidthKey)
    const parsed = Number(raw)
    if (!Number.isFinite(parsed)) return defaultSidebarWidth
    return Math.min(maxSidebarWidth, Math.max(minSidebarWidth, parsed))
  })
  const [sidebarCollapsed, setSidebarCollapsed] = useState(() => localStorage.getItem(sidebarCollapsedKey) === "1")

  const conversationLabel = useMemo(() => {
    if (!activeConversationId) return "New conversation"
    return activeConversationId.slice(0, 8)
  }, [activeConversationId])

  async function handleSend() {
    const clean = prompt.trim()
    if (!clean || busy || historyBusy || clean.length > promptLimit) return

    await onSend(clean, activeConversationId)
    setPrompt("")
  }

  async function handleCopy(content: string) {
    try {
      await navigator.clipboard.writeText(content)
    } catch {
      // no-op
    }
  }

  useEffect(() => {
    localStorage.setItem(sidebarWidthKey, String(sidebarWidth))
  }, [sidebarWidth])

  useEffect(() => {
    localStorage.setItem(sidebarCollapsedKey, sidebarCollapsed ? "1" : "0")
  }, [sidebarCollapsed])

  function handleSidebarResizeStart(event: React.MouseEvent<HTMLDivElement>) {
    if (sidebarCollapsed) return

    event.preventDefault()
    const startX = event.clientX
    const startWidth = sidebarWidth

    function onMouseMove(moveEvent: MouseEvent) {
      const nextWidth = startWidth + (moveEvent.clientX - startX)
      setSidebarWidth(Math.min(maxSidebarWidth, Math.max(minSidebarWidth, nextWidth)))
    }

    function onMouseUp() {
      window.removeEventListener("mousemove", onMouseMove)
      window.removeEventListener("mouseup", onMouseUp)
    }

    window.addEventListener("mousemove", onMouseMove)
    window.addEventListener("mouseup", onMouseUp)
  }

  return (
    <div className="relative flex h-[100dvh] max-h-[100dvh] flex-col overflow-hidden bg-muted/20 p-2 sm:p-4 md:flex-row md:gap-3">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_15%_10%,oklch(0.72_0.14_250_/_0.15),transparent_30%),radial-gradient(circle_at_85%_10%,oklch(0.78_0.12_180_/_0.15),transparent_30%)]" />
      <Card className="hidden min-h-0 md:flex md:flex-col" style={{ width: sidebarCollapsed ? 64 : sidebarWidth }}>
        <CardHeader className={cn("pb-3", sidebarCollapsed && "px-2") }>
          <div className={cn("flex items-center justify-between gap-2", sidebarCollapsed && "justify-center") }>
            {!sidebarCollapsed && (
              <div>
                <CardTitle className="text-lg">Roognis</CardTitle>
                <div className="text-xs text-muted-foreground">Logged in as {user.username}</div>
              </div>
            )}
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => setSidebarCollapsed((previous) => !previous)}
            >
              {sidebarCollapsed ? <PanelLeftOpen className="h-4 w-4" /> : <PanelLeftClose className="h-4 w-4" />}
            </Button>
          </div>
        </CardHeader>
        <CardContent className="flex min-h-0 flex-1 flex-col gap-3 pt-0">
          {!sidebarCollapsed ? (
            <>
              <Button variant="secondary" onClick={() => onSelectConversation(null)}>
                New conversation
              </Button>
              <Separator />
              <ScrollArea className="min-h-0 flex-1 pr-2">
                <div className="space-y-2">
                  {conversations.map((conversation) => (
                    <button
                      key={conversation.id}
                      onClick={() => onSelectConversation(conversation.id)}
                      className={cn(
                        "w-full rounded-md border px-3 py-2 text-left text-sm transition-colors hover:bg-accent",
                        conversation.id === activeConversationId && "bg-accent",
                      )}
                    >
                      <div className="font-medium">{conversation.title || `Conversation ${conversation.id.slice(0, 6)}`}</div>
                      <div className="text-xs text-muted-foreground">Updated {new Date(conversation.updated_at).toLocaleString()}</div>
                    </button>
                  ))}
                  {conversations.length === 0 && <p className="text-sm text-muted-foreground">No conversations yet.</p>}
                </div>
              </ScrollArea>
              <Button variant="outline" onClick={onLogout}>
                <LogOut className="mr-2 h-4 w-4" />
                Logout
              </Button>
            </>
          ) : (
            <>
              <Button type="button" variant="secondary" size="icon" className="mx-auto" onClick={() => onSelectConversation(null)}>
                <Sparkles className="h-4 w-4" />
              </Button>
              <div className="flex-1" />
              <Button type="button" variant="outline" size="icon" className="mx-auto" onClick={onLogout}>
                <LogOut className="h-4 w-4" />
              </Button>
            </>
          )}
        </CardContent>
      </Card>

      <div
        className={cn("relative z-10 hidden w-1 cursor-col-resize rounded-full bg-border/70 transition-colors md:block", sidebarCollapsed && "pointer-events-none opacity-0")}
        onMouseDown={handleSidebarResizeStart}
      />

      <Card className="flex min-h-0 flex-1 flex-col overflow-hidden">
        <CardHeader className="flex-row items-center justify-between space-y-0 border-b pb-4">
          <div className="space-y-1">
            <CardTitle className="flex items-center gap-2 text-xl">
              <Sparkles className="h-5 w-5 text-primary" />
              Inference Workspace
            </CardTitle>
            <div className="text-sm text-muted-foreground">Conversation: {conversationLabel}</div>
          </div>
          <div className="flex items-center gap-3">
            <div className="hidden items-center gap-2 sm:flex">
              <Sun className="h-4 w-4" />
              <Switch checked={theme === "dark"} onCheckedChange={(checked) => onThemeChange(checked ? "dark" : "light")} />
              <Moon className="h-4 w-4" />
            </div>
            <Button
              type="button"
              variant="outline"
              size="icon"
              className="hidden md:inline-flex"
              onClick={() => setSidebarCollapsed((previous) => !previous)}
            >
              {sidebarCollapsed ? <PanelLeftOpen className="h-4 w-4" /> : <PanelLeftClose className="h-4 w-4" />}
            </Button>
            <Badge variant="secondary">{user.role}</Badge>
          </div>
        </CardHeader>

        <CardContent className="flex min-h-0 flex-1 flex-col gap-4 p-3 sm:p-4">
          <div className="space-y-3 rounded-lg border bg-background p-3 md:hidden">
            <div className="text-sm font-medium">Quick controls</div>
            <select
              className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm"
              value={activeConversationId ?? "__new__"}
              onChange={(event) => onSelectConversation(event.target.value === "__new__" ? null : event.target.value)}
            >
              <option value="__new__">New conversation</option>
              {conversations.map((conversation) => (
                <option key={conversation.id} value={conversation.id}>
                  {(conversation.title || `Conversation ${conversation.id.slice(0, 6)}`).slice(0, 42)}
                </option>
              ))}
            </select>
            <div className="flex items-center justify-between rounded-md border px-3 py-2">
              <span className="text-sm">Dark theme</span>
              <Switch checked={theme === "dark"} onCheckedChange={(checked) => onThemeChange(checked ? "dark" : "light")} />
            </div>
            <Button variant="outline" className="w-full" onClick={onLogout}>
              <LogOut className="mr-2 h-4 w-4" />
              Logout
            </Button>
          </div>

          <div className="grid gap-3 rounded-lg border bg-background p-3 md:grid-cols-3">
            <label className="space-y-1 text-sm">
              <span className="text-muted-foreground">Model</span>
              <select
                className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm"
                value={settings.model}
                onChange={(event) => onSettingsChange({ ...settings, model: event.target.value })}
              >
                <option value="qwen2.5:0.5b">qwen2.5:0.5b</option>
                <option value="qwen2.5:1.5b">qwen2.5:1.5b</option>
                <option value="qwen2.5:3b">qwen2.5:3b</option>
              </select>
            </label>
            <label className="space-y-1 text-sm">
              <span className="text-muted-foreground">Temperature ({settings.temperature.toFixed(1)})</span>
              <input
                type="range"
                min={0}
                max={1}
                step={0.1}
                value={settings.temperature}
                onChange={(event) => onSettingsChange({ ...settings, temperature: Number(event.target.value) })}
                className="w-full"
              />
            </label>
            <label className="space-y-1 text-sm">
              <span className="text-muted-foreground">Max tokens</span>
              <div className="relative">
                <Gauge className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <input
                  type="number"
                  min={64}
                  max={4096}
                  step={64}
                  value={settings.maxTokens}
                  onChange={(event) => onSettingsChange({ ...settings, maxTokens: Number(event.target.value) || 1024 })}
                  className="h-10 w-full rounded-md border border-input bg-background pl-9 pr-3 text-sm"
                />
              </div>
            </label>
          </div>

          <ScrollArea className="min-h-0 flex-1 rounded-lg border bg-background p-4">
            <div className="space-y-4">
              {historyBusy && (
                <div className="flex items-center gap-2 rounded-md border bg-muted/30 px-3 py-2 text-sm text-muted-foreground">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Loading conversation history...
                </div>
              )}
              {messages.map((message) => (
                <div
                  key={message.id}
                  className={cn(
                    "flex gap-3 rounded-lg border p-3",
                    message.role === "assistant" ? "bg-secondary/35" : "bg-background",
                  )}
                >
                  <div className="mt-0.5">
                    {message.role === "assistant" ? <Bot className="h-4 w-4" /> : <UserIcon className="h-4 w-4" />}
                  </div>
                  <div className="flex-1 space-y-2">
                    <MarkdownContent content={message.content} />
                    <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                      <span>{new Date(message.createdAt).toLocaleTimeString()}</span>
                      {message.model && <Badge variant="outline">{message.model}</Badge>}
                      {typeof message.latencyMs === "number" && <Badge variant="outline">{message.latencyMs.toFixed(0)} ms</Badge>}
                      {message.cached && <Badge variant="secondary">cached</Badge>}
                      <Button variant="ghost" size="sm" className="h-6 px-2 text-xs" onClick={() => handleCopy(message.content)}>
                        <Copy className="mr-1 h-3 w-3" />
                        Copy
                      </Button>
                    </div>
                  </div>
                </div>
              ))}
              {messages.length === 0 && <p className="text-sm text-muted-foreground">Ask your first question to start.</p>}
            </div>
          </ScrollArea>

          <div className="space-y-3">
            <Textarea
              value={prompt}
              onChange={(event) => setPrompt(event.target.value.slice(0, promptLimit))}
              onKeyDown={(event) => {
                if (event.key !== "Enter" || event.shiftKey || busy || historyBusy) return
                if (event.nativeEvent.isComposing) return
                event.preventDefault()
                void handleSend()
              }}
              placeholder="Ask anything..."
              className="min-h-[110px]"
            />
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <p className={cn("break-words text-xs", prompt.length > promptLimit * 0.9 ? "text-amber-500" : "text-muted-foreground")}>
                Prompt: {prompt.length.toLocaleString()}/{promptLimit.toLocaleString()} • Enter to send • Shift+Enter for newline
              </p>
              <Button
                className="self-end shrink-0 sm:self-auto"
                onClick={handleSend}
                disabled={busy || historyBusy || prompt.trim().length === 0 || prompt.trim().length > promptLimit}
              >
                {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
                <span className="ml-2">Send</span>
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

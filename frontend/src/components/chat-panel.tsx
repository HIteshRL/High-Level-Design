import { useMemo, useState } from "react"
import { Bot, Loader2, LogOut, Send, User as UserIcon } from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Separator } from "@/components/ui/separator"
import { Textarea } from "@/components/ui/textarea"
import type { Conversation, InferenceResponse, User } from "@/lib/types"
import { cn } from "@/lib/utils"

interface ChatPanelProps {
  user: User
  conversations: Conversation[]
  activeConversationId: string | null
  messages: Array<{ role: "user" | "assistant"; content: string }>
  busy: boolean
  onSelectConversation: (id: string | null) => void
  onSend: (prompt: string, conversationId: string | null) => Promise<InferenceResponse>
  onLogout: () => void
}

export function ChatPanel({
  user,
  conversations,
  activeConversationId,
  messages,
  busy,
  onSelectConversation,
  onSend,
  onLogout,
}: ChatPanelProps) {
  const [prompt, setPrompt] = useState("")

  const conversationLabel = useMemo(() => {
    if (!activeConversationId) return "New conversation"
    return activeConversationId.slice(0, 8)
  }, [activeConversationId])

  async function handleSend() {
    const clean = prompt.trim()
    if (!clean || busy) return

    await onSend(clean, activeConversationId)
    setPrompt("")
  }

  return (
    <div className="grid min-h-screen grid-cols-1 bg-muted/20 p-4 md:grid-cols-[280px_1fr] md:gap-4">
      <Card className="hidden md:flex md:flex-col">
        <CardHeader className="pb-3">
          <CardTitle className="text-lg">Roognis</CardTitle>
          <div className="text-xs text-muted-foreground">Logged in as {user.username}</div>
        </CardHeader>
        <CardContent className="flex flex-1 flex-col gap-3 pt-0">
          <Button variant="secondary" onClick={() => onSelectConversation(null)}>
            New conversation
          </Button>
          <Separator />
          <ScrollArea className="h-[calc(100vh-280px)] pr-2">
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
        </CardContent>
      </Card>

      <Card className="flex min-h-[90vh] flex-col">
        <CardHeader className="flex-row items-center justify-between space-y-0 border-b pb-4">
          <div className="space-y-1">
            <CardTitle className="text-xl">Inference Workspace</CardTitle>
            <div className="text-sm text-muted-foreground">Conversation: {conversationLabel}</div>
          </div>
          <Badge variant="secondary">{user.role}</Badge>
        </CardHeader>

        <CardContent className="flex flex-1 flex-col gap-4 p-4">
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
            <Button variant="outline" className="w-full" onClick={onLogout}>
              <LogOut className="mr-2 h-4 w-4" />
              Logout
            </Button>
          </div>

          <ScrollArea className="h-[58vh] rounded-lg border bg-background p-4">
            <div className="space-y-4">
              {messages.map((message, index) => (
                <div
                  key={`${message.role}-${index}`}
                  className={cn(
                    "flex gap-3 rounded-lg border p-3",
                    message.role === "assistant" ? "bg-secondary/35" : "bg-background",
                  )}
                >
                  <div className="mt-0.5">
                    {message.role === "assistant" ? <Bot className="h-4 w-4" /> : <UserIcon className="h-4 w-4" />}
                  </div>
                  <p className="text-sm leading-6 whitespace-pre-wrap">{message.content}</p>
                </div>
              ))}
              {messages.length === 0 && <p className="text-sm text-muted-foreground">Ask your first question to start.</p>}
            </div>
          </ScrollArea>

          <div className="space-y-3">
            <Textarea
              value={prompt}
              onChange={(event) => setPrompt(event.target.value)}
              placeholder="Ask anything..."
              className="min-h-[110px]"
            />
            <div className="flex items-center justify-between">
              <p className="text-xs text-muted-foreground">Prompt limit: 32,000 Unicode characters.</p>
              <Button onClick={handleSend} disabled={busy || prompt.trim().length === 0}>
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

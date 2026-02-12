import type {
  Conversation,
  ConversationMessage,
  InferenceRequest,
  InferenceResponse,
  TokenResponse,
  User,
} from "@/lib/types"

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080"

class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.status = status
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
  })

  const isJson = response.headers.get("content-type")?.includes("application/json")
  const payload = isJson ? await response.json() : null

  if (!response.ok) {
    const message = payload?.message ?? payload?.error ?? `Request failed: ${response.status}`
    throw new ApiError(message, response.status)
  }

  return payload as T
}

export async function register(username: string, email: string, password: string): Promise<TokenResponse> {
  return request<TokenResponse>("/api/v1/auth/register", {
    method: "POST",
    body: JSON.stringify({ username, email, password }),
  })
}

export async function login(username: string, password: string): Promise<TokenResponse> {
  return request<TokenResponse>("/api/v1/auth/token", {
    method: "POST",
    body: JSON.stringify({ username, password }),
  })
}

export async function me(token: string): Promise<User> {
  return request<User>("/api/v1/auth/me", {
    method: "GET",
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function listConversations(token: string): Promise<Conversation[]> {
  return request<Conversation[]>("/api/v1/conversations", {
    method: "GET",
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function listConversationMessages(token: string, conversationId: string): Promise<ConversationMessage[]> {
  const query = new URLSearchParams({ conversation_id: conversationId })
  return request<ConversationMessage[]>(`/api/v1/conversation-messages?${query.toString()}`, {
    method: "GET",
    headers: { Authorization: `Bearer ${token}` },
  })
}

export async function complete(token: string, body: InferenceRequest): Promise<InferenceResponse> {
  return request<InferenceResponse>("/api/v1/inference/complete", {
    method: "POST",
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify({ ...body, stream: false }),
  })
}

export { ApiError, API_BASE }

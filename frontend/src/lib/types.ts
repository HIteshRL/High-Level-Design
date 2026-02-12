export type UserRole = "student" | "teacher" | "parent" | "admin"

export interface User {
  id: string
  username: string
  email: string
  full_name?: string | null
  role: UserRole
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface TokenResponse {
  access_token: string
  token_type: string
  expires_in: number
}

export interface Conversation {
  id: string
  user_id: string
  title?: string | null
  created_at: string
  updated_at: string
}

export interface InferenceRequest {
  prompt: string
  conversation_id?: string
  stream?: boolean
  model?: string
  temperature?: number
  max_tokens?: number
}

export interface InferenceResponse {
  id: string
  conversation_id: string
  content: string
  model: string
  token_count?: number
  latency_ms: number
  cached: boolean
}

import { createContext, useContext, useState, useCallback } from 'react'

const AuthContext = createContext(null)

const TOKEN_KEY = 'fb_token'
const USER_KEY  = 'fb_user'

export function AuthProvider({ children }) {
  const [token, setToken]   = useState(() => localStorage.getItem(TOKEN_KEY) || null)
  const [user, setUser]     = useState(() => localStorage.getItem(USER_KEY)  || null)

  const login = useCallback((newToken, username) => {
    localStorage.setItem(TOKEN_KEY, newToken)
    localStorage.setItem(USER_KEY,  username)
    setToken(newToken)
    setUser(username)
  }, [])

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
    setToken(null)
    setUser(null)
  }, [])

  // Authenticated fetch helper — automatically attaches Bearer token
  const apiFetch = useCallback(async (url, options = {}) => {
    const headers = {
      ...(options.headers || {}),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    }
    // Don't set Content-Type for FormData (browser sets it with boundary)
    if (!(options.body instanceof FormData)) {
      headers['Content-Type'] = headers['Content-Type'] || 'application/json'
    }
    const resp = await fetch(url, { ...options, headers })
    if (resp.status === 401) {
      logout()
      throw new Error('session_expired')
    }
    return resp
  }, [token, logout])

  return (
    <AuthContext.Provider value={{ token, user, login, logout, apiFetch, isLoggedIn: !!token }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => useContext(AuthContext)

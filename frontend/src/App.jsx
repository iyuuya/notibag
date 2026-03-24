import { useState, useEffect, useRef } from 'react'
import './App.css'

function App() {
  const [notifications, setNotifications] = useState([])
  const [hidingNotifications, setHidingNotifications] = useState(new Set())
  const [currentTime, setCurrentTime] = useState(new Date())
  const [connectionStatus, setConnectionStatus] = useState('connecting') // 'connected' | 'disconnected' | 'connecting'
  const ws = useRef(null)

  useEffect(() => {
    let reconnectTimer = null
    let reconnectDelay = 1000
    const maxReconnectDelay = 30000
    let intentionallyClosed = false

    // WebSocket接続
    const connectWebSocket = () => {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/ws`

      ws.current = new WebSocket(wsUrl)

      ws.current.onopen = () => {
        console.log('WebSocket connected')
        setConnectionStatus('connected')
        reconnectDelay = 1000
        // 初期通知データを要求
        ws.current.send(JSON.stringify({ type: 'get_notifications' }))
      }

      ws.current.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          if (data.type === 'notification') {
            setNotifications(prev => [data.notification, ...prev])
          } else if (data.type === 'notifications_list') {
            setNotifications(data.notifications || [])
          }
        } catch (error) {
          console.error('Failed to parse message:', error)
        }
      }

      ws.current.onclose = () => {
        console.log('WebSocket disconnected')
        if (!intentionallyClosed) {
          setConnectionStatus('disconnected')
          console.log(`Reconnecting in ${reconnectDelay}ms...`)
          reconnectTimer = setTimeout(() => {
            setConnectionStatus('connecting')
            reconnectDelay = Math.min(reconnectDelay * 2, maxReconnectDelay)
            connectWebSocket()
          }, reconnectDelay)
        }
      }

      ws.current.onerror = (error) => {
        console.error('WebSocket error:', error)
      }
    }

    connectWebSocket()

    return () => {
      intentionallyClosed = true
      clearTimeout(reconnectTimer)
      if (ws.current) {
        ws.current.close()
      }
    }
  }, [])

  // 時間を1秒ごとに更新
  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(new Date())
    }, 1000)

    return () => clearInterval(timer)
  }, [])

  const markAsRead = (id) => {
    // 即座にスライドアウトアニメーションを開始
    setHidingNotifications(prev => new Set([...prev, id]))
    
    // WebSocketで既読マーク（バックグラウンドで）
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({ 
        type: 'mark_read', 
        notification_id: id 
      }))
    }
    
    // アニメーション完了後に通知をリストから削除
    setTimeout(() => {
      setNotifications(prev => prev.filter(notif => notif.id !== id))
      setHidingNotifications(prev => {
        const newSet = new Set(prev)
        newSet.delete(id)
        return newSet
      })
    }, 300)
  }

  const formatRelativeTime = (timestamp) => {
    const now = currentTime
    const notificationTime = new Date(timestamp)
    const diffInSeconds = Math.floor((now - notificationTime) / 1000)

    if (diffInSeconds < 60) {
      return `${diffInSeconds}秒前`
    } else if (diffInSeconds < 3600) {
      const minutes = Math.floor(diffInSeconds / 60)
      return `${minutes}分前`
    } else if (diffInSeconds < 86400) {
      const hours = Math.floor(diffInSeconds / 3600)
      return `${hours}時間前`
    } else {
      const days = Math.floor(diffInSeconds / 86400)
      return `${days}日前`
    }
  }

  const formatCurrentDateTime = () => {
    const now = currentTime
    const date = now.toLocaleDateString('ja-JP', {
      month: 'numeric',
      day: 'numeric',
      weekday: 'short'
    })
    const time = now.toLocaleTimeString('ja-JP', {
      hour: '2-digit',
      minute: '2-digit'
    })
    return { date, time }
  }

  const { date, time } = formatCurrentDateTime()

  return (
    <div className="App">
      {connectionStatus !== 'connected' && (
        <div className={`connection-banner ${connectionStatus}`}>
          {connectionStatus === 'disconnected' ? '接続が切れました。再接続を待っています...' : '再接続中...'}
        </div>
      )}
      <div className="clock-widget">
        <div className="clock-time">{time}</div>
        <div className="clock-date">{date}</div>
      </div>

      <main className="notifications-container">
        {notifications.length === 0 ? (
          <div className="empty-state">
            <div className="empty-icon">🔔</div>
            <p>通知はありません</p>
          </div>
        ) : (
          <div className="notifications-grid">
            {notifications.map((notification) => (
              <div 
                key={notification.id} 
                className={`notification-card unread type-${notification.type || 'info'} ${hidingNotifications.has(notification.id) ? 'hiding' : ''}`}
                onClick={() => !hidingNotifications.has(notification.id) && markAsRead(notification.id)}
                onTouchStart={() => {}} // タッチ反応を改善
              >
                <div className="notification-header">
                  <span className="notification-title">{notification.title}</span>
                  <span className="notification-time">{formatRelativeTime(notification.timestamp)}</span>
                </div>
                <div className="notification-body">
                  {notification.message}
                </div>
              </div>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}

export default App
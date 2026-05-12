import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import { SagaProvider } from './context/SagaContext' // Import cái file ông vừa tạo
import './index.css'

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <SagaProvider>
      <App />
    </SagaProvider>
  </React.StrictMode>,
)
import { Routes, Route, Navigate } from 'react-router-dom'
import { Layout } from './components/Layout'
import { PrivateRoute } from './components/PrivateRoute'
import { Login } from './pages/Login'
import { Catalog } from './pages/catalog.jsx'
import { BeerDetail } from './pages/BeerDetail'
import { Promotions } from './pages/Promotions'
import { OrderInfo } from './pages/OrderInfo'
import { Contacts } from './pages/Contacts'
import { AdminParams } from './pages/AdminParams'
import { AdminEnums } from './pages/AdminEnums'
import { AdminCategories } from './pages/AdminCategories'
import { AdminAggregates } from './pages/AdminAggregates'
import { Toast } from './components/Toast'
import { ToastProvider } from './hooks/useToast'

function App() {
    return (
        <ToastProvider>
            <Routes>
                <Route path="/login" element={<Login />} />
                <Route path="/" element={<Layout />}>
                    <Route index element={<Catalog />} />
                    <Route path="catalog" element={<Catalog />} />
                    <Route path="beer/:id" element={<BeerDetail />} />
                    <Route path="promotions" element={<Promotions />} />
                    <Route path="order-info" element={<OrderInfo />} />
                    <Route path="contacts" element={<Contacts />} />

                    <Route path="admin" element={<PrivateRoute><AdminParams /></PrivateRoute>} />
                    <Route path="admin/params" element={<PrivateRoute><AdminParams /></PrivateRoute>} />
                    <Route path="admin/enums" element={<PrivateRoute><AdminEnums /></PrivateRoute>} />
                    <Route path="admin/categories" element={<PrivateRoute><AdminCategories /></PrivateRoute>} />
                    <Route path="admin/aggregates" element={<PrivateRoute><AdminAggregates /></PrivateRoute>} />
                </Route>
                <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
            <Toast />
        </ToastProvider>
    )
}

export default App
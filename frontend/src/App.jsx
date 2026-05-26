import { Routes, Route, Navigate } from 'react-router-dom'
import { Catalog } from './pages/catalog'
import { BeerDetail } from './pages/beerDetail'
import { Promotions } from './pages/Promotions'
import { OrderInfo } from './pages/OrderInfo'
import { Contacts } from './pages/Contacts'
import { Login } from './pages/Login'
import { AdminPanel } from './pages/AdminPanel'
import { AdminParams } from './pages/AdminParams'
import { AdminEnums } from './pages/AdminEnums'
import { AdminCategories } from './pages/AdminCategories'
import { AdminAggregates } from './pages/AdminAggregates'
import { AdminBeers } from './pages/AdminBeers'
import { Layout } from './components/Layout'
import { PrivateRoute } from './components/PrivateRoute'
import { ToastProvider } from './hooks/useToast'
import { Toast } from './components/Toast'

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
                    <Route path="admin" element={<PrivateRoute><AdminPanel /></PrivateRoute>}>
                        <Route index element={<Navigate to="beers" replace />} />
                        <Route path="beers" element={<AdminBeers />} />
                        <Route path="params" element={<AdminParams />} />
                        <Route path="aggregates" element={<AdminAggregates />} />
                        <Route path="categories" element={<AdminCategories />} />
                        <Route path="enums" element={<AdminEnums />} />
                    </Route>
                </Route>
                <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
            <Toast />
        </ToastProvider>
    )
}

export default App
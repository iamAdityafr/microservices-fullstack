import { BrowserRouter, Routes, Route } from "react-router-dom";
import Home from "./pages/Home";
import Cart from "./pages/Cart";
import ProductDetails from "./pages/ProductDetailsModal";
import PrivateRoute from "./components/PrivateRoute";
import Checkout from "./pages/Checkout";
import Success from "./pages/Success";
import Cancel from "./pages/Cancel";
import Search from "./pages/Search";
import Register from "./pages/Register";
function App() {
  return (
    <div className="relative min-h-screen w-full">
      <div
        className="absolute inset-0 -z-10"
        style={{
          background: `linear-gradient(225deg, #FFB3D9 0%, #FFD1DC 20%, #FFF0F5 40%, #E6F3FF 60%, #D1E7FF 80%, #C7E9F1 100%)`,
        }}
      />

      <Routes>
          // public
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<Register />} />

        // TODO: move "/success" to protected routes
        <Route path="/success" element={<Success />} />

          // protected
        <Route element={<PrivateRoute />}>
          <Route path="/cart" element={<Cart />} />
          <Route path="/product/:id" element={<ProductDetails />} />
          <Route path="/checkout/:cartId" element={<Checkout />} />
          <Route path="/cancel" element={<Cancel />} />
          <Route path="/search" element={<Search />} />

        </Route>

      </Routes>
    </div>
  );
}

export default App;

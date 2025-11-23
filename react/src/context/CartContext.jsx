import { createContext, useContext, useEffect, useState } from "react";
import { useAuth } from "./AuthContext";
import api from "../api";

const CartContext = createContext();

export const useCart = () => useContext(CartContext);

export const CartProvider = ({ children }) => {
    const [cart, setCart] = useState(null);
    const [cartItems, setCartItems] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    const { user } = useAuth();

    const addToCart = async (product) => {
        if (!user?.id) return;

        try {
            await api.post(`/cart/add`, { product_id: product.id });

            setCartItems((prev) => {
                if (prev.some((item) => item.product_id === product.id)) {
                    return prev;
                }
                return [...prev, product];
            });
        } catch (err) {
            console.error("Failed to add cart:", err);
        }
    };

    const removeFromCart = async (productId) => {
        try {
            await api.delete(`/cart/remove`, { data: { product_id: productId } });
            setCartItems((prev) => prev.filter((item) => item.product_id !== productId));
        } catch (err) {
            console.error("Failed remove from cart:", err);
        }
    };

    useEffect(() => {
        const fetchCart = async () => {
            if (!user?.id) {
                console.log('No user ID skip cart fetch');
                setLoading(false);
                return;
            }

            try {
                console.log('Fetching cart for user:', user.id);
                console.log('call /cart/getcart');

                const response = await api.get(`/cart/getcart`);
                console.log('Cart resp:', response);

                if (response.data && Array.isArray(response.data.items)) {
                    setCart(response.data);

                    setCartItems(response.data.items);
                    console.log('Cart items:', response.data.items);
                } else {
                    setError('Unexp API resp struct');
                }
            } catch (err) {
                console.error('Load cart error');
                console.error('error body :', err)
                console.error('err resp:', err.response);
                console.error('err status:', err.response?.status);
                console.error('err data:', err.response?.data);
                setError('couldnt load the cart');
            } finally {
                setLoading(false);
            }
        };

        fetchCart();
    }, [user?.id]);

    return (
        <CartContext.Provider value={{ cart, cartItems, loading, error, setCartItems, removeFromCart, addToCart }}>
            {children}
        </CartContext.Provider>
    );
};
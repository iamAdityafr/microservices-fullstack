import { useCart } from "../context/CartContext";
import Header from "../components/Header";
import Loading from "../components/Loading";
import { useNavigate } from "react-router-dom";

// i am doing localhost for uploads rn 
const IMGURL = "http://localhost:8082/uploads/"
const Cart = () => {
    const { cartItems, loading, error, removeFromCart, cart } = useCart();
    if (loading) return <Loading />;
    if (error) return <p>Sm went wrong.</p>;
    if (!cartItems || cartItems.length === 0) return (
        <div className="text-2xl p-2 bg-gray-500 text-center">Cart Empty</div>

    );
    return (
        <>
            <Header />
            <div className="p-4 sm:p-8 max-w-7xl mx-auto">
                <h1 className="text-4xl font-bold mb-8 text-gray-800 text-center">Your Shopping Cart</h1>
                <div className="flex flex-col lg:flex-row gap-10">
                    <CartAside cartItems={cartItems} cart={cart} />

                    <div className="flex-1 space-y-6 bg-white p-6 md:p-8 rounded-xl shadow-lg">
                        {cartItems.map((item) => {
                            const imgSrc = item.image?.startsWith("uploads/") ? item.image.slice(8) : item.image;
                            const itemPrice = (item.price_cents || 0) / 100;
                            return (
                                <div key={item.product_id}
                                    className="flex items-start justify-between border-b last:border-b-0 pb-6 pt-2">
                                    <div className="flex items-start gap-4">
                                        <img src={`${IMGURL}${imgSrc}`} alt={item.name} className="w-24 h-24 object-cover rounded-lg shadow-md" />
                                        <div>
                                            <h2 className="font-semibold text-xl text-gray-800">{item.name}</h2>
                                            <button onClick={() => removeFromCart(item.product_id)} className="mt-3 text-sm text-red-600 font-medium hover:text-red-800 transition duration-150"> Remove </button>
                                        </div>
                                    </div>

                                    <div className="text-right flex flex-col items-end">
                                        <p className="font-bold text-lg text-gray-900">${(itemPrice * (item.quantity || 1)).toFixed(2)}</p>

                                        <div className="flex items-center gap-0 mt-3 border border-gray-300 rounded-lg overflow-hidden divide-x divide-gray-300">
                                        </div>
                                    </div>
                                </div>
                            );
                        })}
                    </div>

                </div>
            </div>
        </>
    );
};

export default Cart;


const CartAside = ({ cartItems, cart }) => {
    const totalItems = cartItems.reduce((sum, item) => sum + (item.quantity || 1), 0);
    const totalPriceCents = cartItems.reduce((sum, item) => sum + (item.price_cents || 0) * (item.quantity || 1), 0);
    const totalPrice = totalPriceCents / 100;
    const navigate = useNavigate();

    const handleCheckoutClick = () => {
        const cartId = cart?.id || cart?.user_id;
        if (cartId) {
            navigate(`/checkout/${cartId}`);
        } else {
            console.error("No cart ID available");
            alert("Unable to proceed to checkout. Please try refreshing the page.");
        }
    }

    return (
        <div className="w-full lg:w-1/3 bg-white p-8 rounded-xl shadow-2xl h-fit sticky top-4">
            <h2 className="text-2xl font-bold mb-4 text-gray-800">Order Summary</h2>
            <hr className="mb-6 border-gray-200" />

            <div className="space-y-3 mb-6">
                {cartItems.map((item) => {
                    const itemPrice = (item.price_cents || 0) / 100;
                    return (
                        <div key={item.product_id} className="flex justify-between text-base text-gray-600">
                            <span> {item.name} <span className="text-sm font-bold"></span>
                            </span>
                            <span>${(itemPrice * (item.quantity || 1)).toFixed(2)}</span>
                        </div>)
                })}
            </div>

            <hr className="mb-4 border-dashed border-black" />

            <p className="font-medium text-lg text-gray-700">Total items: <span className="font-semibold">{totalItems}</span></p>
            <div className="flex justify-between items-center mt-3 pt-3 border-double border-black">
                <p className="font-bold"> Order Total: </p>
                <p className="font-extrabold text-3xl"> ${totalPrice.toFixed(2)} </p>
            </div>

            <button className="mt-6 w-full bg-blue-600 text-white text-lg font-semibold py-3 rounded-lg shadow-lg hover:bg-blue-700 transition duration-200 hover:scale-[1.06]" onClick={handleCheckoutClick}> Proceed </button>
        </div>
    );
};
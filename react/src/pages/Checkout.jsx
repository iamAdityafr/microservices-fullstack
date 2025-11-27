import React, { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { loadStripe } from "@stripe/stripe-js";
import { Elements, PaymentElement, useStripe, useElements } from "@stripe/react-stripe-js";
import api from "../api";

const stripePKey = import.meta.env.VITE_VITE_STIPE_PK2;
const stripePromise = loadStripe(`${stripePKey}`);

function CheckoutForm() {
    const stripe = useStripe();
    const elements = useElements();
    const [message, setMessage] = useState("");
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!stripe || !elements) return;

        setLoading(true);

        const { error } = await stripe.confirmPayment({
            elements,
            confirmParams: {
                return_url: `${window.location.origin}/success`
            },
        });

        if (error) {
            setMessage(` ${error.message}`);
        } else if (paymentIntent) {
            switch (paymentIntent.status) {
                case "succeeded":
                    setMessage("Payment succeeded!");
                    break;
                case "processing":
                    setMessage("Payment is processing...");
                    break;
                case "requires_payment_method":
                    setMessage("Payment failed");
                    break;
                default:
                    setMessage("Sm happened.");
            }
        }
        setLoading(false);
    };

    return (
        <form onSubmit={handleSubmit} className="space-y-6">
            <div className="p-4 border border-gray-300 rounded-lg bg-white shadow-sm">
                <PaymentElement />
            </div>
            <button
                type="submit"
                disabled={!stripe || loading}
                className="w-full bg-blue-500 text-white py-3 px-4 rounded-lg font-semibold transition duration-300 scale-[1.05] hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed">
                {loading ? "Processing..." : "Pay Now"}
            </button>
            {message && (
                <div
                    className="p-3 rounded-xl text-center bg-gray-200 text-black">
                    {message}
                </div>
            )}
        </form>
    );
}

export default function Checkout() {
    const { cartId } = useParams();
    const navigate = useNavigate();
    const [cart, setCart] = useState(null);
    const [clientSecret, setClientSecret] = useState("");
    const [error, setError] = useState("");

    useEffect(() => {
        async function fetchCartAndPaymentIntent() {
            try {
                const cartResponse = await api.get(`/cart/getcart`);
                if (!cartResponse.data || !Array.isArray(cartResponse.data.items) || cartResponse.data.items.length === 0) {
                    setError("Your cart is empty");
                    return;
                }

                setCart(cartResponse.data);
                const totalAmount = cartResponse.data.items.reduce(
                    (sum, item) => sum + item.price_cents * item.quantity, 0);

                const paymentData = await api.post("/payments/intent", {
                    order_id: cartResponse.data.id || cartResponse.data.user_id,
                    amount: totalAmount,
                    currency: "usd",
                });

                if (paymentData.data.client_secret) {
                    setClientSecret(paymentData.data.client_secret);
                } else {
                    throw new Error("No client secret received");
                }
            } catch (err) {
                setError(err.message || "Failed to load checkout");
            }
        }

        fetchCartAndPaymentIntent();
    }, [cartId]);

    if (error) {
        return (
            <CheckoutError />
        );
    }

    if (!cart || !clientSecret) {
        return (
            <div className="min-h-screen flex items-center justify-center">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-e-4 text-bold border-green-400 mx-auto mb-4"></div>
                    <p className="text-black font-semibold">Loading checkout...</p>
                </div>
            </div>
        );
    }

    const totalAmount = cart.items.reduce(
        (sum, item) => sum + item.price_cents * item.quantity,
        0
    );

    const options = {
        clientSecret,
        appearance: {
            theme: "stripe",
            variables: {
                colorPrimary: "#345CD9",
                fontFamily: "Monospace",
                borderRadius: "8px",
            },
        },
    };

    return (
        <div className="min-h-screen p-4 sm:p-8 bg-gray-100">
            <div className="shadow-3xl mx-auto max-w-5xl rounded-2xl bg-white p-10 sm:p-8">
                <h1 className="text-3xl font-bold mb-6 text-gray-800">Checkout</h1>
                <hr class="p-5" />
                <div className="mb-8 p-4 bg-green-100 rounded-lg border-dashed">
                    <h2 className="text-xl font-semibold mb-4 text-gray-800">Order Summary</h2>
                    <ul className="space-y-3 mb-4">
                        {cart.items.map((item) => (
                            <li key={item.product_id} className="flex justify-between text-gray-700">
                                <span>
                                    {item.name} <span className="text-gray-500">x {item.quantity}</span>
                                </span>
                                <span className="font-medium">
                                    ${((item.price_cents * item.quantity) / 100).toFixed(2)}
                                </span>
                            </li>
                        ))}
                    </ul>
                    <span className="text-lg font-bold text-gray-800">Total:</span> <br />
                    <span className="text-2xl font-bold text-blue-600">
                        ${(totalAmount / 100).toFixed(2)}
                    </span>
                </div>

                <div>
                    <h2 className="text-xl font-semibold mb-4 text-gray-800">Payment Details</h2>
                    <Elements stripe={stripePromise} options={options}>
                        <CheckoutForm clientSecret={clientSecret} />
                    </Elements>
                </div>

            </div>
            <button
                onClick={() => navigate("/cart")}
                className="mt-6 scale-[1.10] transform bg-blue-200 p-2 font-medium text-blue-600 duration-200" >
                Back to Cart
            </button>
        </div>
    );
}
const CheckoutError = () => {
    return (
        <div
            className="p-3 rounded-lg text-center">{message}
        </div>
    )
}
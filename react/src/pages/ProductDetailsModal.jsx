import { useCart } from "../context/CartContext";

const ProductDetailsModal = ({ product, onClose }) => {
    const { addToCart } = useCart();

    if (!product) return null;

    return (
        <div
            className="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4"
            onClick={onClose}>
            <div className="relative bg-white rounded-xl shadow-2xl w-full max-w-5xl h-[85vh] overflow-hidden" onClick={(e) => e.stopPropagation()}>
                <button
                    onClick={onClose}
                    className="absolute top-4 right-4 z-10 p-2 text-white bg-blue-500 rounded-full shadow-lg text-2xl font-extrabold hover:bg-blue-600 transition duration-300">
                    Close
                </button>

                <div className="flex flex-col md:flex-row h-full">
                    <div className="md:w-1/2 p-8 flex flex-col justify-between overflow-y-auto">
                        <div className="space-y-6 mt-12">
                            <div>
                                <h3 className="text-gray-500 font-medium text-sm uppercase mb-1">Name</h3>
                                <h2 className="text-4xl font-extrabold text-gray-900">{product.name}</h2>
                            </div>
                            <hr class="m-2 border-dashed mt-12 mb-7"></hr>
                            <div>
                                <h3 className="text-gray-500 font-medium text-sm uppercase mb-1">Description</h3>
                                <p className="text-gray-700 text-base leading-relaxed">{product.description}</p>
                            </div>
                            <hr class="m-2 border-dashed mt-7 mb-7"></hr>
                            <div>
                                <h3 className="text-gray-500 font-medium text-sm uppercase mb-1">Price</h3>
                                <p className="text-3xl font-bold text-black">${product.price}</p>
                            </div>
                        </div>

                        <button
                            onClick={() => {
                                addToCart(product);
                                onClose();
                            }} className="w-full bg-blue-600 text-white font-semibold py-2 rounded-lg shadow-lg hover:bg-blue-700 transition duration-300 transform hover:scale-[1.01] mt-6" >Add to Cart</button>
                    </div>
                       <div className="md:w-1/2 p-6 flex justify-center items-center bg-gray-50">
                        <img src={product.image} alt={product.name} className="w-full h-full max-h-[calc(85vh-3rem)] object-contain rounded-xl" />
                    </div>
                </div>
            </div>
        </div>
    );
};

export default ProductDetailsModal;

import { useCart } from "../context/CartContext";
import ProductDetailsModal from "../pages/ProductDetailsModal";
import { useState } from "react";
import { createPortal } from "react-dom";
const UPLOADS = import.meta.env.VITE_IMG_UPLOADS;

const ProductCard = ({ product }) => {
  const { addToCart } = useCart();
  const [showModal, setShowModal] = useState(false);

  if (!product) return <p>No product data</p>;
  const { image, name, description, price } = product;
  const imgSrc = image.startsWith("uploads/") ? image.slice(8) : image;
  return (
    <>
      <div onClick={() => setShowModal(true)} className=" w-72 cursor-pointer border-1 border-black backdrop-blur-xl rounded-md p-4  shadow-2xl transition-transform duration-300">
        <img src={`${UPLOADS}${imgSrc}`} alt={name} className="w-full h-60 rounded-2xl shadow-lg object-cover mb-4" />
        <h3 className="text-lg font-bold mb-1">Name: {name}</h3>
        <p className="text-black mb-1">Description: {description}</p>
        <p className="text-black mb-1">Price: {price}</p>
        <button onClick={(e) => {
          e.stopPropagation();
          addToCart(product);
        }} className="bg-blue-500 p-2 w-full rounded-lg transition shadow-lg duration-200 ease-in-out transform hover:scale-[1.05] mt-2">Add to cart</button>
      </div>

      {showModal && createPortal(
        <ProductDetailsModal product={product} onClose={() => setShowModal(false)}/>,
        document.body )}
    </>
  );
};

export default ProductCard;
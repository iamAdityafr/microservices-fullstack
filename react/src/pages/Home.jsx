import Header from "../components/Header";
import ProductCard from "../components/ProductCard";
import Loading from "../components/Loading";
import { useProduct } from "../context/ProductContext";

const Home = () => {
  const { products, loading, error } = useProduct();

  return (
    <>
      <Header />
      <div className="p-6 sm:p-8 max-w-7xl mx-auto">
        {loading && <Loading />}
        {error && <p>Sm went wrong</p> }
        {!loading && !error && (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 sm:gap-6 place-items-center">
            {products.map((product) => (
              <ProductCard key={product.id} product={product} />
            ))}
          </div>
        )}
      </div>
    </>
  );
};

export default Home;

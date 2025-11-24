import React, { useState } from 'react';
import Header from '../components/Header';
import ProductCard from '../components/ProductCard';
import api from '../api';
import Loading from '../components/Loading';

const Search = () => {
    const [query, setQuery] = useState('');
    const [results, setResults] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const handleSearch = async () => {
        if (!query) return;
        setLoading(true);
        setError(null);

        try {
            const res = await api.get(`/products/search?q=${encodeURIComponent(query)}`);
            setResults(res.data || []);
        } catch (err) {
            console.error(err);
            setError(err.response?.data?.message || err.message || 'couldnt fetch products');
        } finally {
            setLoading(false);
        }
    };


    const handleKeyPress = (e) => {
        if (e.key === 'Enter') handleSearch();
    };

    return (
        <>
            <Header />

            <div className="min-h-screen flex justify-center p-6 ">
                <div className="w-full max-w-7xl text-center backdrop-blur-md bg-white/30 rounded-2xl p-6">
                    <h1 className="text-3xl font-bold mb-6">Search Products</h1>
                    <div className="flex gap-3 mb-8">
                        <input
                            type="text"
                            value={query}
                            onChange={(e) => setQuery(e.target.value)}
                            onKeyDown={handleKeyPress}
                            placeholder="Enter product here"
                            className="flex-1 p-3 border-2 border-transparent bg-white/20 rounded-xl shadow-lg focus:ring-blue-500"/>
                        <button onClick={handleSearch} className="px-6 py-3 bg-blue-500 text-white rounded-xl shadow hover:bg-blue-600 transition duration-300 hover:scale-[1.07] ">Search</button>
                    </div>

                    {loading && <Loading/>}
                    {error && <p>err: {error}</p>}

                    {results.length === 0 && !loading && <p>No results found</p>}

                    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">{results.map((product) => (
                            <ProductCard key={product.id} product={product} />
                        ))}
                    </div>
                </div>
            </div>
        </>
    );
};

export default Search;

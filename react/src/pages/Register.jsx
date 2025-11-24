import React, { useState, useEffect } from 'react';
import { TiHome } from 'react-icons/ti';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from "../context/AuthContext";

const AuthCard = () => {
    const { login, logout, user } = useAuth();
    const [formData, setFormData] = useState({email: '',password: '',confirmPassword: '',});
    const navigate = useNavigate();
    const [isFlipped, setIsFlipped] = useState(false);
    const handleChange = (e) => { 
        const { name, value } = e.target; setFormData((prev) => ({...prev, [name]: value, })); 
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        if (form.password !== form.confirmPassword) {
            setError("Passwords don't match");
            return;
        }
        setLoading(true);
        try {
            await register({
                name: form.name,
                email: form.email,
                password: form.password,
            });
            navigate('/');
        } catch (err) {
            setError(err.response?.data?.message || 'Registration failed');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (user) {
            navigate('/', { replace: true });
        }
    }, [user, navigate]);
    return user ? (
        <div className="flex justify-center items-center min-h-screen text-black">
            <div className="text-center">
                <h1 className="text-2xl font-bold text-black mb-4">You are already logged in!</h1>
                <button onClick={logout} className="px-6 py-2 bg-red-500 hover:bg-red-600 text-black rounded-lg">Click to logout </button>
            </div>
        </div>
    ) : (
        <>
            <div className="flex justify-center items-center min-h-screen p-12 text-black"
                style={{ background: 'linear-gradient(225deg, #FFB3D9 0%, #FFD1DC 20%, #FFF0F5 40%, #E6F3FF 60%, #D1E7FF 80%, #C7E9F1 100%)', }}>
                    
                <div className="w-full max-w-md relative" style={{ perspective: '1000px' }}>

                    <div className="w-full h-[500px] relative transition-transform duration-320" style={{transformStyle: 'preserve-3d',transform: isFlipped ? 'rotateY(180deg)' : 'rotateY(0deg)',}}>

                            {/* Register */}
                        <div className="absolute inset-0 w-full h-full p-8 rounded-2xl border-2 border-black shadow-xl backdrop-blur-xl" style={{backfaceVisibility: 'hidden',transformStyle: "preserve-3d",background: 'rgba(255,255,255,0.15)',}}>
                            <div className="absolute top-4 right-4">
                                
                                <button onClick={() => navigate('/')} className="text-black transition duration-300  p-3 rounded-3xl scale-[1.10] bg-blue-300">
                                    <TiHome size={24} className="transition duration-300 hover:scale-[1.15]" />
                                </button>
                            </div>
                            <h1 className="text-2xl font-bold mb-8 text-center text-black drop-shadow-lg">Register your profile</h1>

                            <form onSubmit={handleSubmit} className="space-y-4">
                                <div className="space-y-1">
                                    <label className="block text-black/90 font-medium text-sm mb-1">Your email</label>
                                    <input
                                        type="email"
                                        name="email"
                                        value={formData.email}
                                        onChange={handleChange}
                                        placeholder="e@mail.com"
                                        className="w-full px-4 py-3 bg-white/20 border placeholder:text-gray border-gray-300 rounded-xl text-black focus:outline-none focus:ring-2 focus:ring-white/30 focus:placeholder-transparent"/>
                                </div>

                                <div className="space-y-1">
                                    <label className="block text-black/90 font-medium text-sm mb-1">Password</label>
                                    <input
                                        type="password"
                                        name="password"
                                        value={formData.password}
                                        onChange={handleChange}
                                        placeholder="••••••••"
                                        className="w-full px-4 py-3 bg-white/20 border placeholder:text-gray border-gray-300 rounded-xl text-black focus:outline-none focus:ring-2 focus:ring-white/30 focus:placeholder-transparent"/>
                                </div>

                                <div className="space-y-1">
                                    <label className="block text-black/90 font-medium text-sm mb-1">Confirm password</label>
                                    <input
                                        type="password"
                                        name="confirmPassword"
                                        value={formData.confirmPassword}
                                        onChange={handleChange}
                                        placeholder="••••••••"
                                        className="w-full px-4 py-3 bg-white/20 border placeholder:text-gray border-gray-300 rounded-xl text-black focus:outline-none focus:ring-2 focus:ring-white/30 focus:placeholder-transparent" />
                                </div>

                                <button type="submit" className="w-full py-4 bg-blue-500 hover:bg-blue-600 transition duration-300 hover:scale-[1.02] font-medium rounded-xl shadow-lg hover:shadow-xl mt-2">Create an account
                                </button>
                            </form>

                            <p className="text-center text-sm text-black mt-3">
                                Already have an account?{' '}
                                <span className="font-medium hover:underline cursor-pointer " onClick={() => setIsFlipped(true)}> Login here </span>
                            </p>
                        </div>

                        {/* Login */}
                        <div className="absolute inset-0 w-full h-full p-8 rounded-2xl border-2 border-black shadow-2xl text-black backdrop-blur-xl"
                            style={{backfaceVisibility: 'hidden', transform: 'rotateY(180deg)',transformStyle: "preserve-3d",background: 'rgba(255,255,255,0.15)',}}>
                            <div className="absolute top-4 right-4">
                                <button onClick={() => navigate('/')} className="text-black transition duration-300 hover:scale-[1.07] p-3 rounded-3xl bg-blue-300">
                                    <TiHome size={24} className="transition duration-300 hover:scale-[1.15]" />
                                </button>
                            </div>
                            <h1 className="text-2xl font-bold mb-8 text-center drop-shadow-lg">Login</h1>
                            <form
                                className="space-y-4"
                                onSubmit={(e) => {
                                    e.preventDefault();
                                    login({ email: formData.email, password: formData.password,});}}>
                                <div className="space-y-1">
                                    <label className="block text-black font-medium text-sm mb-1">Email</label>
                                    <input
                                        type="email"
                                        name="email"
                                        value={formData.email}
                                        onChange={handleChange}
                                        placeholder="e@mail.com"
                                        className="w-full px-4 py-3 bg-white/20 border border-gray-300 rounded-xl placeholder-gray text-black focus:outline-none focus:ring-white/30 focus:placeholder-transparent"
                                        required />
                                </div>

                                <div className="space-y-1">
                                    <label className="block text-black font-medium text-sm mb-1">Password</label>
                                    <input
                                        type="password"
                                        name="password"
                                        value={formData.password}
                                        onChange={handleChange}
                                        placeholder="••••••••"
                                        className="w-full px-4 py-3 bg-white/20 border border-gray-300 rounded-xl placeholder-gray text-black focus:outline-none focus:ring-white/30 focus:placeholder-transparent"
                                        required/>
                                </div>

                                <button
                                    type="submit"
                                    className="w-full py-4 bg-green-500 hover:bg-green-600 transition duration-300 hover:scale-[1.05] font-medium rounded-xl shadow-lg hover:shadow-xl">
                                    Login
                                </button>
                            </form>


                            <p className="text-center text-sm text-black mt-6">
                                Don't have an account?{' '}
                                <span className="font-bold text-black hover:underline cursor-pointer" onClick={() => setIsFlipped(false)}> Register here
                                </span>
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </>
    );
};

export default AuthCard;



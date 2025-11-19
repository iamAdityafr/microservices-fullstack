import { createContext, useContext, useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../api';

const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);
    const navigate = useNavigate();

    useEffect(() => {
        const checkUser = async () => {
            try {
                const { data } = await api.get('/profile');
                setUser(data);
            } catch (err) {
                setUser(null);
            } finally {
                setLoading(false);
            }
        };

        checkUser();
    }, []);

    const login = async (credentials) => {
        try {
            await api.post('/login', credentials);
            const { data } = await api.get('/profile');
            setUser(data);
            navigate('/');
        } catch (err) {
            alert('Login failed');
        }
    };

    const logout = async () => {
        try {
            await api.post('/logout'); 
        } catch (err) {
            console.error('Logout error', err);
        } finally {
            setUser(null);
            navigate('/login');
        }
    };

    return (
        <AuthContext.Provider value={{ user, login, logout, loading }}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => useContext(AuthContext);

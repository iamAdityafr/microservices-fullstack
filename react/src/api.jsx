import axios from "axios";

const port = import.meta.env.VITE_GATEWAY_PORT
console.log(port)
const api = axios.create({
    baseURL: port,
    withCredentials: true,
});

export default api;

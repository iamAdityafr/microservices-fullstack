import { Link } from "react-router-dom";
import { TbSearch } from "react-icons/tb";
import { useAuth } from "../context/AuthContext";

const Header = () => {
  const { user } = useAuth();

  return (
    <div className="justify-center flex text-white font-mono pt-4 backdrop-blur-md mb-6">
      <div className="flex justify-between items-center p-2 bg-black w-5/6 rounded-2xl">
        <Link to="/" className="text-xl font-bold ml-3 shadow-lg transition duration-300 ease-in-out transform hover:scale-[1.20]">LOGO</Link>
        <div className="flex gap-2">
          <Link
            to="/search"
            className="inline-flex mx-3 p-2 bg-fuchsia-500 rounded-lg shadow-lg transition duration-200 ease-in-out transform hover:scale-[1.10] hover:bg-fuchsia-600 font-bold "
          >
            <TbSearch size={20} />
          </Link>
          <Link
            to="/cart"
            className="inline-flex mx-3 p-2 bg-fuchsia-500 rounded-lg shadow-lg transition duration-200 ease-in-out transform hover:scale-[1.10] hover:bg-fuchsia-600 font-bold "
          >
            Cart
          </Link>
          {
            user ? (
              <h1 className="mt-2 m-2">hello here</h1>
            ) : (
              <Link
                to="/login"
                className="inline-flex mx-3 p-2 bg-fuchsia-500 rounded-lg shadow-lg transition duration-200 ease-in-out transform hover:scale-[1.10] hover:bg-fuchsia-600 font-bold "
              >
                Login
              </Link>
            )
          }

        </div>
      </div>
    </div>
  );
};

export default Header;
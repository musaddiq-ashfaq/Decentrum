import React from "react";
import { Link } from "react-router-dom";
import { Home, PlusCircle, User } from 'lucide-react';

const Navbar = () => {
  return (
    <nav className="bg-white shadow-md">
      <div className="container mx-auto px-4">
        <div className="flex justify-between items-center py-4">
          <Link to="/" className="text-2xl font-bold text-[#052a47]">Decentrum</Link>
          <ul className="flex space-x-6">
            <li>
              <Link to="/feed" className="flex items-center text-gray-600 hover:text-[#4dbf38]">
                <Home className="h-5 w-5 mr-1" />
                <span>Feed</span>
              </Link>
            </li>
            <li>
              <Link to="/post" className="flex items-center text-gray-600 hover:text-[#4dbf38]">
                <PlusCircle className="h-5 w-5 mr-1" />
                <span>Create Post</span>
              </Link>
            </li>
            <li>
              <Link to="/profile" className="flex items-center text-gray-600 hover:text-[#4dbf38]">
                <User className="h-5 w-5 mr-1" />
                <span>Profile</span>
              </Link>
            </li>
          </ul>
        </div>
      </div>
    </nav>
  );
};

export default Navbar;


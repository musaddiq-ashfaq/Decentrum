// Navbar.js
import React from "react";
import { Link } from "react-router-dom"; // Use Link for navigation
import "./Navbar.css"; // Import CSS for styling

const Navbar = () => {
  return (
    <nav className="navbar">
      <ul>
        <li>
          <Link to="/feed">Feed</Link>
        </li>
        <li>
          <Link to="/post">Create Post</Link>
        </li>
       
      </ul>
    </nav>
  );
};

export default Navbar;

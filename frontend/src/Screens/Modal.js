// src/components/Modal.js
import React from "react";
import "./Modal.css"; // Create a CSS file for modal styles

const Modal = ({ message, onClose }) => {
  return (
    <div className="modal-overlay">
      <div className="modal">
        <p>{message}</p>
        <button onClick={onClose} className="modal-button">
          OK
        </button>
      </div>
    </div>
  );
};

export default Modal;

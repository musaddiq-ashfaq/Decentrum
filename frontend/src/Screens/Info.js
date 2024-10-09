import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom'; // Import useNavigate instead
import './UserInfo.css'; // Import the CSS file

const UserInfo = () => {
    const [userInfo, setUserInfo] = useState({
        name: '',
        email: '',
        phone: '',
        profilePicture: null,
    });

    const navigate = useNavigate(); // Initialize navigate

    const handleChange = (e) => {
        const { name, value, files } = e.target;
        if (name === 'profilePicture') {
            setUserInfo({
                ...userInfo,
                [name]: files[0], // Capture the file for profile picture
            });
        } else {
            setUserInfo({
                ...userInfo,
                [name]: value,
            });
        }
    };

    const handleSubmit = async (e) => {
        e.preventDefault();

        const formData = new FormData();
        formData.append('name', userInfo.name);
        formData.append('email', userInfo.email);
        formData.append('phone', userInfo.phone);
        if (userInfo.profilePicture) {
            formData.append('profilePicture', userInfo.profilePicture);
        }

        try {
            const response = await fetch('http://localhost:8081/info', { // Ensure the correct endpoint
                method: 'POST',
                body: formData,
            });

            if (response.ok) {
                console.log('User information submitted successfully');
                navigate('/login'); // Redirect to login page using navigate
            } else {
                console.error('Failed to submit user information', response.status);
                // Handle errors (e.g., show a message to the user)
            }
        } catch (error) {
            console.error('Error submitting user information:', error);
            // Handle network or other errors
        }
    };

    return (
        <div className="user-info-container">
            <h1>Complete Your Profile</h1>
            <form onSubmit={handleSubmit} className="user-info-form">
                <label>
                    Name:
                    <input
                        type="text"
                        name="name"
                        value={userInfo.name}
                        onChange={handleChange}
                        required
                    />
                </label>
                <label>
                    Email:
                    <input
                        type="email"
                        name="email"
                        value={userInfo.email}
                        onChange={handleChange}
                        required
                    />
                </label>
                <label>
                    Phone:
                    <input
                        type="tel"
                        name="phone"
                        value={userInfo.phone}
                        onChange={handleChange}
                        required
                    />
                </label>
                <label>
                    Profile Picture:
                    <input
                        type="file"
                        name="profilePicture"
                        accept="image/*"
                        onChange={handleChange}
                    />
                </label>
                <button type="submit" className="submit-button">Submit</button>
            </form>
        </div>
    );
};

export default UserInfo;

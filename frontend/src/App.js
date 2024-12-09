import { Route, Routes } from 'react-router-dom';
import ChatApp from './Screens/Chat';
import Feed from './Screens/Feed';
import Info from './Screens/Info';
import LoginPage from "./Screens/Login";
import Post from './Screens/Post';
import RegistrationPage from "./Screens/SignUp";
const App = () => {
  return (
    <div>
      <Routes>
        <Route path="/" element={<LoginPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/signup" element={<RegistrationPage />} />
        <Route path="/info" element={<Info />} />
        <Route path="/post" element={<Post />} />
        {/* <Route path="/reaction" element={<Feed />} /> */}
        <Route path="/feed" element={<Feed />} />
        <Route path="/chat" element={<ChatApp />} />
      </Routes>
    </div>
  );
};

export default App;

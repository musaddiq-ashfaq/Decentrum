import { Route, Routes } from 'react-router-dom';
import ChatApp from './Screens/Chat';
import Feed from './Screens/Feed';
import Info from './Screens/Info';
import LoginPage from './Screens/Login';
import Post from './Screens/Post';
import RegistrationPage from './Screens/SignUp';
import CreateGroup from './Screens/Groups';
import UsersList from './Screens/UsersList';
import FriendRequests from './Screens/FriendRequests';
import FriendsList from './Screens/FriendsList';

const App = () => {
  return (
    <div>
      <Routes>
        <Route path="/" element={<LoginPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/signup" element={<RegistrationPage />} />
        <Route path="/info" element={<Info />} />
        <Route path="/post" element={<Post />} />
        <Route path="/feed" element={<Feed />} />
        <Route path="/chat" element={<ChatApp />} />
        <Route path="/group" element={<CreateGroup />} />
        <Route path="/users" element={<UsersList />} />
        <Route path="/friend-requests" element={<FriendRequests />} />
        <Route path="/friends" element={<FriendsList />} />
      </Routes>
    </div>
  );
};

export default App;
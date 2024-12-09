import { NavigationMenuLink, NavigationMenu } from "./navigation-menu";
import { AvatarImage, AvatarFallback, Avatar } from "./avatar";
import { Route } from "react-router-dom";
import { Bell, MessageCircle } from "lucide-react";
import homepageScreen from "../../Screens/homepageScreen";
import RecentChatsScreen from "../../Screens/recentChatsScreen";
import ChatScreen from "../../Screens/chatScreen";
import loginScreen from "../../Screens/loginScreen";
import { Link } from "react-router-dom/cjs/react-router-dom";

export default function Navbar() {
  return (
    <div className="flex h-screen">
      <NavigationMenu className="bg-gray-900 text-gray-400 p-4 flex flex-col justify-between">
        <div className="space-y-4">
          <NavigationMenuLink asChild>
            <Link
              className="flex items-center gap-3 rounded-md p-2 hover:bg-gray-800"
               to="/rchats"
            >
              <MessageCircle className="h-6 w-6" />
              <span className="sr-only">Chat</span>
            </Link>
          </NavigationMenuLink>
          <NavigationMenuLink asChild>
            <Link
              className="flex items-center gap-3 rounded-md p-2 hover:bg-gray-800"
               to="/rchats"
            >
              <Bell className="h-6 w-6" />
              <span className="sr-only">Notifications</span>
            </Link>
          </NavigationMenuLink>
          <NavigationMenuLink asChild>
            <Link
              className="flex items-center gap-3 rounded-md p-2 hover:bg-gray-800"
               to="/rchats"
            >
              <UserPlusIcon className="h-6 w-6" />
              <span className="sr-only">Add Friends</span>
            </Link>
          </NavigationMenuLink>
        </div>
        <div className="space-y-4">
          <NavigationMenuLink asChild>
            <Link
              className="flex items-center gap-3 rounded-md p-2 hover:bg-gray-800"
               to="/rchats"
            >
              <SettingsIcon className="h-6 w-6" />
              <span className="sr-only">Settings</span>
            </Link>
          </NavigationMenuLink>
          <NavigationMenuLink asChild>
            <Link
              className="flex items-center gap-3 rounded-md p-2 hover:bg-gray-800"
               to="/rchats"
            >
              <Avatar className="h-6 w-6">
                <AvatarImage alt="@shadcn" src="/placeholder-avatar.jpg" />
                <AvatarFallback>JP</AvatarFallback>
              </Avatar>
              <span className="sr-only">Profile</span>
            </Link>
          </NavigationMenuLink>
        </div>
      </NavigationMenu>
      <main className="flex-1 p-8">
        <Route path="/" component={homepageScreen} exact />
        <Route path="/rchats" component={RecentChatsScreen} exact />
        <Route path="/chat" component={ChatScreen} exact />
        <Route path="/login" component={loginScreen} />
      </main>
    </div>
  );
}

function SettingsIcon(props) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  );
}

function UserPlusIcon(props) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
      <circle cx="9" cy="7" r="4" />
      <line x1="19" x2="19" y1="8" y2="14" />
      <line x1="22" x2="16" y1="11" y2="11" />
    </svg>
  );
}

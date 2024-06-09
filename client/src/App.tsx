import { useState } from "react";
import "./App.css";
import NavBar from "./components/NavBar";
import Planer from "./components/Planer";

export type Task = {
  id: number;
  title: string;
  desc: string;
  isDone: boolean;
  category: string;
};
export type User = {
  name: string;
  password: string;
  tasks: Task[];
};
export const BASE_URL = "http://localhost:5000/api";
function App() {
  const [currentUser, setCurrentUser] = useState<User>({
    name: "",
    password: "",
    tasks: [],
  });
  return (
    <div>
      <NavBar user={currentUser} setUser={setCurrentUser} />
      <Planer user={currentUser} setUser={setCurrentUser} />
    </div>
  );
}

export default App;

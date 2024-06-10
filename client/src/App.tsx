import { useState } from "react";
import "./App.css";
import NavBar from "./components/NavBar";
import Planer from "./components/Planer";

export type Task = {
  id: number;
  title: string;
  desc: string;
  isDone: boolean;
  category: Category;
};
export type User = {
  name: string;
  tasks: Task[];
  categories: Category[];
};
export type Category = {
  cat_name: string;
  color_header: string;
  color_body: string;
};
export const BASE_URL = "http://localhost:5000/api";
function App() {
  const [user, setUser] = useState<User>({
    name: "",
    tasks: [],
    categories: [],
  });
  return (
    <div>
      <NavBar user={user} setUser={setUser} />
      {user.name !== "" && <Planer user={user} setUser={setUser} />}
      {user.name === "" && (
        <h1>Melde dich an, um deine Aufgaben zu sehen und zu verwalten</h1>
      )}
    </div>
  );
}

export default App;

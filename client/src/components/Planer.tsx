import { useState } from "react";
import ToDoForm from "./ToDoForm";
import ToDo from "./ToDo";
import { Button } from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import { Task } from "../App";
import "../App.css";

export default function Planer(props: any) {
  const [showForm, setShowForm] = useState(false);
  const todoElements = props.user.tasks.map((task: Task, idx: number) => (
    <ToDo key={idx} data={task} setUser={props.setUser} />
  ));
  return (
    <div>
      <ToDoForm
        open={showForm}
        setShowForm={setShowForm}
        setUser={props.setUser}
        type="create"
      />
      <Button
        onClick={() => setShowForm(true)}
        variant="contained"
        className="todo--button"
      >
        <AddIcon /> Aufgabe erstellen
      </Button>
      {todoElements}
    </div>
  );
}

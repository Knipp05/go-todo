import { useState } from "react";
import ToDoForm from "./ToDoForm";
import ToDo from "./ToDo";
import { Button } from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import { Task } from "../App";

export default function Planer(props: any) {
  const [showForm, setShowForm] = useState(false);
  const todoElements = props.user.tasks.map((task: Task, idx: number) => (
    <ToDo key={idx} data={task} />
  ));
  return (
    <div>
      <ToDoForm open={showForm} onClose={setShowForm} />
      <Button onClick={() => setShowForm(true)} variant="contained">
        <AddIcon /> Aufgabe erstellen
      </Button>
      {todoElements}
    </div>
  );
}

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
    <ToDo
      key={idx}
      data={task}
      user={props.user}
      setUser={props.setUser}
      categories={props.user.categories}
    />
  ));
  return (
    <div>
      <Button
        onClick={() => setShowForm(true)}
        variant="contained"
        sx={{ marginTop: 2, marginLeft: "45%" }}
      >
        <AddIcon /> Aufgabe erstellen
      </Button>
      <ToDoForm
        open={showForm}
        setShowForm={setShowForm}
        setUser={props.setUser}
        user={props.user}
        categories={props.user.categories}
        type="create"
      />
      {todoElements}
    </div>
  );
}

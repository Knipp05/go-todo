import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  TextField,
} from "@mui/material";
import { useState } from "react";
export default function ToDoForm(props: any) {
  const [taskContent, setTaskContent] = useState({ title: "", desc: "" });
  function handleInput(event: any) {
    setTaskContent((oldContent) => {
      return { ...oldContent, [event.target.name]: event.target.value };
    });
  }
  const submitInput = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    // TODO
  };
  return (
    <Dialog
      open={props.open}
      onClose={() => props.onClose(false)}
      PaperProps={{
        component: "form",
        onSubmit: (event: React.FormEvent<HTMLFormElement>) =>
          submitInput(event),
      }}
    >
      <DialogContent>
        <DialogTitle>Neue Aufgabe anlegen</DialogTitle>
        <DialogContentText>
          Bitte gib einen Titel und optional eine Beschreibung für die Aufgabe
          an
        </DialogContentText>
        <TextField
          required
          id="title"
          name="title"
          label="Titel"
          type="text"
          variant="standard"
          onChange={handleInput}
        />
        <TextField
          id="desc"
          name="desc"
          label="Beschreibung"
          type="text"
          variant="standard"
          onChange={handleInput}
        />
        <DialogActions>
          <Button type="submit">Aufgabe erstellen</Button>
        </DialogActions>
      </DialogContent>
    </Dialog>
  );
}

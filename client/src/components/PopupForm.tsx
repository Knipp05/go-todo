import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  TextField,
} from "@mui/material";
import "../App.css";
import { useState } from "react";
import { BASE_URL } from "../App";

export default function PopupForm(props: any) {
  const [userCredentials, setUserCredentials] = useState({
    name: "",
    password: "",
  });
  const [repeatedPassword, setRepeatedPassword] = useState("");
  const [errorMessage, setErrorMessage] = useState("");

  function handleInput(event: any) {
    if (event.target.name === "password_repeat") {
      setRepeatedPassword(event.target.value);
    } else {
      setUserCredentials((oldCredentials) => {
        return { ...oldCredentials, [event.target.name]: event.target.value };
      });
    }
  }

  const submitInput = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (props.showPopup === 1) {
      if (userCredentials.password !== repeatedPassword) {
        setErrorMessage("Passwörter sind nicht identisch!");
        return;
      }

      try {
        const res = await fetch(BASE_URL + `/users/new`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(userCredentials),
        });

        const data = await res.json();

        if (!res.ok) {
          throw new Error(data.error || "Unbekannter Fehler aufgetreten");
        }

        setUserCredentials({ name: "", password: "" });
        setRepeatedPassword("");
        setErrorMessage("");
        props.setShowPopup(0);
      } catch (error: any) {
        setErrorMessage(error.message);
      }
    }
    if (props.showPopup === 2) {
      try {
        const res = await fetch(BASE_URL + `/users`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(userCredentials),
        });

        const data = await res.json();

        if (!res.ok) {
          throw new Error(data.error || "Unbekannter Fehler aufgetreten");
        }

        setUserCredentials({ name: "", password: "" });
        setRepeatedPassword("");
        setErrorMessage("");
        sessionStorage.setItem("token", data.token);
        props.setUser({
          name: data.name,
          tasks: data.tasks,
          categories: data.categories,
        });
        props.setShowPopup(0);
      } catch (error: any) {
        setErrorMessage(error.message);
      }
    }
  };
  return (
    <Dialog
      open={props.showPopup > 0 ? true : false}
      onClose={() => {
        setErrorMessage("");
        props.setShowPopup(0);
      }}
      PaperProps={{
        component: "form",
        onSubmit: (event: React.FormEvent<HTMLFormElement>) =>
          submitInput(event),
      }}
    >
      <DialogTitle>
        {props.showPopup === 1
          ? "Neuen Benutzer registrieren"
          : "Mit bestehendem Benutzer anmelden"}
      </DialogTitle>
      <DialogContent className="popup--content">
        <TextField
          required
          id="name"
          name="name"
          label="Benutzername"
          type="text"
          variant="standard"
          onChange={handleInput}
        />
        <TextField
          required
          id="password"
          name="password"
          label="Passwort"
          type="password"
          variant="standard"
          onChange={handleInput}
        />
        {props.showPopup === 1 && (
          <TextField
            required
            id="password_repeat"
            name="password_repeat"
            label="Passwort wiederholen"
            type="password"
            variant="standard"
            onChange={handleInput}
          />
        )}
        {errorMessage && (
          <DialogContentText sx={{ color: "red" }}>
            {errorMessage}
          </DialogContentText>
        )}
        <DialogActions>
          <Button type="submit">
            {props.showPopup === 1 ? "registrieren" : "anmelden"}
          </Button>
        </DialogActions>
      </DialogContent>
    </Dialog>
  );
}

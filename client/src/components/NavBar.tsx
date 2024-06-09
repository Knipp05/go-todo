import { Button } from "@mui/material";
import { useState } from "react";
import PopupForm from "./PopupForm";
import { BASE_URL } from "../App";

export default function NavBar(props: any) {
  const [showPopup, setShowPopup] = useState(0);
  function togglePopup(popupId: number) {
    setShowPopup(popupId);
  }
  const handleLogout = async () => {
    const token = sessionStorage.getItem("token");
    if (token) {
      try {
        const res = await fetch(BASE_URL + `/users/logout`, {
          method: "DELETE",
          headers: {
            "Content-Type": "application/json",
            Authorization: token,
          },
        });

        const data = await res.json();

        if (!res.ok) {
          throw new Error(data.error || "Unbekannter Fehler aufgetreten");
        }
        props.setUser({ name: "", password: "", tasks: [] });
        sessionStorage.removeItem("token");
      } catch (error: any) {
        throw new Error("Logout nicht erfolgreich");
      }
    }
  };
  return (
    <header className="navbar">
      <h1>
        {props.user.name === "" ? "ToDo Planer" : `Hallo ${props.user.name}`}
      </h1>
      <div>
        {props.user.name === "" && (
          <Button variant="contained" onClick={() => togglePopup(1)}>
            Registrieren
          </Button>
        )}
        {props.user.name === "" && (
          <Button variant="contained" onClick={() => togglePopup(2)}>
            Login
          </Button>
        )}
        {props.user.name !== "" && (
          <Button variant="contained" onClick={handleLogout}>
            Logout
          </Button>
        )}
      </div>
      <PopupForm
        showPopup={showPopup}
        togglePopup={togglePopup}
        setUser={props.setUser}
      />
    </header>
  );
}

module Signup exposing (main)

import Html exposing (..)
import Html.Attributes exposing (..)
import Browser
import Html.Events exposing(onInput)

type alias User =
    { username : String
    , email : String
    , password : String
    , passwordAgain : String
    , loggedIn : Bool
    }

-- MODEL

initialModel : User
initialModel =
    { username = ""
    , email = ""
    , password = ""
    , passwordAgain = ""
    , loggedIn = False
    }

-- UPDATE

type Msg
    = Username String
    | Email String
    | Password String
    | PasswordAgain String
    --| LoggedIn Bool

update : Msg -> User -> User
update msg model = 
    case msg of 
        Username username ->
            { model | username = username}

        Email email ->
            { model | email = email}

        Password password ->
            { model | password = password}

        PasswordAgain passwordAgain ->
            { model | passwordAgain = passwordAgain}

-- VIEW

view : User -> Html Msg
view model =
    div []
        [ h1 [] [ text "Sign up" ]
        , Html.form []
            [ div []
                [ text "Username"
                , input [ id "username", type_ "text", value model.username, onInput Username ] []
                ]
            , div []
                [ text "Email"
                , input [ id "email", type_ "email", value model.email, onInput Email ] []
                ]
            , div []
                [ text "Password"
                , input [ id "password", type_ "password", value model.password, onInput Password ] []
                ]
            , div []
                [ text "Re-enter Password"
                , input [ id "passwordAgain", type_ "password", value model.passwordAgain, onInput PasswordAgain ] []
                ]
                , viewValidation model
            , div []
                [ button [ type_ "submit" ]
                    [ text "Create my account" ]
                ]
            ] 
        ]

viewValidation : User -> Html msg
viewValidation model =
  if model.password == model.passwordAgain then
    div [] [ text " " ]
  else
    div [ style "color" "red" ] [ text "Passwords do not match!" ]

--main : Html msg
main =
    Browser.sandbox 
    {
        init = initialModel
        ,view = view
        ,update = update
    }
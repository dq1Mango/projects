use tokio::sync::mpsc::UnboundedSender;

use crossterm::event::{self, Event, EventStream, KeyCode};

use futures::{StreamExt, future::FutureExt};

use crate::logger::Logger;
use crate::*;

#[derive(PartialEq)]
pub enum LinkingAction {
  Url(Url),
  Success,
  Fail,
}

#[derive(PartialEq)]
pub enum Action {
  Type(char),
  Backspace,
  Scroll(i16),

  SetMode(Mode),
  SetFocus(Focus),

  Link(LinkingAction),

  Quit,
}
/// Convert Event to Action
///
/// We don't need to pass in a `model` to this function in this example
/// but you might need it as your project evolves
///
/// (the project evolved (pokemon core))
pub async fn handle_crossterm_events(tx: UnboundedSender<Action>, mode: &Arc<Mutex<Mode>>) {
  let mut reader = EventStream::new();

  loop {
    let event = reader.next().fuse().await;
    match event {
      Some(Ok(event)) => match event {
        Event::Key(key) => {
          if key.kind == event::KeyEventKind::Press {
            if let Some(action) = handle_key(key, mode) {
              let err = tx.send(action);
            }
          }
        }
        _ => {}
      },
      Some(Err(err)) => Logger::log(format!("Error reading event: {err} ")),
      None => Logger::log(format!("I dont think this should ever happend")),
    }
  }
}

pub fn handle_key(key: event::KeyEvent, mode: &Arc<Mutex<Mode>>) -> Option<Action> {
  match *mode.lock().unwrap() {
    Mode::Insert => match key.code {
      KeyCode::Esc => Some(Action::SetMode(Mode::Normal)),
      KeyCode::Char(char) => Some(Action::Type(char)),
      // this will not get confusing trust
      KeyCode::Backspace => Some(Action::Backspace),
      _ => None,
    },
    Mode::Normal => match key.code {
      KeyCode::Char('j') => Some(Action::Scroll(-1)),
      KeyCode::Char('k') => Some(Action::Scroll(1)),
      KeyCode::Char('d') => Some(Action::Scroll(-10)),
      KeyCode::Char('u') => Some(Action::Scroll(10)),

      KeyCode::Char('i') => Some(Action::SetMode(Mode::Insert)),
      KeyCode::Char('S') => Some(Action::SetFocus(Focus::Settings)),
      KeyCode::Char('C') => Some(Action::SetFocus(Focus::Chats)),
      // KeyCode::Char('k') => Some(Action::Decrement),
      KeyCode::Char('q') => Some(Action::Quit),
      _ => None,
    },
  }
}

pub fn update(model: &mut Model, msg: Action, logger: &mut Logger) -> Option<Action> {
  match msg {
    Action::Type(char) => {
      model.current_chat().text_input.insert_char(char, logger);
    }
    Action::Backspace => model.current_chat().text_input.delete_char(),

    Action::Scroll(lines) => model.current_chat().location.requested_scroll = lines,

    Action::SetMode(new_mode) => *model.mode.lock().unwrap() = new_mode,

    Action::SetFocus(new_focus) => model.focus = new_focus,

    Action::Quit => {
      // You can handle cleanup and exit here
      // -- im ok thanks tho
      model.running_state = RunningState::OhShit;
    }

    _ => {}
  };

  None
}

// pub fn scroll(chat: &mut Chat, lines: i16, settings: Settings) {
//   let mut lines = lines;
//
//   // oh man i sure hope this ugly line of repeated code will not f*** me over in the future
//     let message_width: u16 = (area.width as f32 * settings.message_width_ratio + 0.5) as u16 - 2;
//
//   loop {
//     let height = chat.messages[chat.location.index].body.as_lines(message_width).len();
//     if lines
//   }
// }

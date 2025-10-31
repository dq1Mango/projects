use crate::logger::Logger;
use crate::*;
use color_eyre;
use crossterm::event::{self, Event, KeyCode};

#[derive(PartialEq)]
pub enum Action {
  Type(char),
  Backspace,
  Scroll(i16),

  SetMode(Mode),

  Quit,
}
/// Convert Event to Action
///
/// We don't need to pass in a `model` to this function in this example
/// but you might need it as your project evolves
///
/// (the project evolved (pokemon core))
pub fn handle_event(mode: &Arc<Mutex<Mode>>) -> color_eyre::Result<Option<Action>> {
  if event::poll(Duration::from_millis(250))? {
    if let Event::Key(key) = event::read()? {
      if key.kind == event::KeyEventKind::Press {
        return Ok(handle_key(key, mode));
      }
    }
  }
  Ok(None)
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

    Action::Quit => {
      // You can handle cleanup and exit here
      // -- im ok thanks tho
      model.running_state = RunningState::OhShit;
    }
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

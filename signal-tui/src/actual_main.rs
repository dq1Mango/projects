#[cfg(test)]
mod tests;

use std::fs::File;
use std::io::Write;
use std::{time::Duration, vec};

use ratatui::layout::{Constraint, Direction, Layout};
use ratatui::{
  Frame,
  buffer::Buffer,
  crossterm::event::{self, Event, KeyCode},
  layout::Rect,
  style::Stylize,
  symbols::border,
  text::Line,
  widgets::{Block, Paragraph, Widget},
};
use std::fs::OpenOptions;

#[derive(Debug, Default)]
struct Model {
  running_state: RunningState,
  mode: Mode,
  chats: Vec<Chat>,
  chat_index: usize,
}

#[derive(Debug, Default, PartialEq, Eq)]
enum RunningState {
  #[default]
  Running,
  OhShit,
}

#[derive(Default, Debug, PartialEq)]
pub enum Mode {
  #[default]
  Normal,
  Insert,
}

// impl PartialEq for Mode {
//   fn eq(&self, other: &Self) -> bool {}
// }

#[derive(PartialEq)]
enum Action {
  Type(char),
  Backspace,

  SetMode(Mode),

  Quit,
}

#[derive(Debug, Default, Clone)]
pub struct MulitLineString {
  body: String,
  cached_lines: Vec<String>,
  cached_width: u16,
  cached_length: u16,
}

impl MulitLineString {
  fn init(str: &str) -> Self {
    Self {
      body: str.to_string(),
      cached_lines: vec!["".to_string()],
      cached_width: 0,
      cached_length: 0,
    }
  }

  fn update_cache(&mut self, width: u16) {
    let mut lines: Vec<String> = Vec::new();
    let mut new_line = String::from("");

    // collumn index
    let mut coldex = 0;
    // let availible_width = (term_width as f32 * settings.message_width_ratio + 0.5) as usize;
    let availible_width = width as usize;

    for yap in self.body.split(" ") {
      let mut length = yap.len();

      if coldex + yap.len() <= availible_width {
        new_line.push_str(yap);
        new_line.push_str(" ");
        coldex += yap.len() + 1;
      } else {
        // INCOMPLETE LOGIC!!! should probably trim the start of the string
        if new_line != "" {
          lines.push(new_line.clone().trim_end().to_string());
        }

        let mut index = 0;

        while length >= availible_width {
          lines.push(yap[index..index + availible_width].to_string());
          length -= availible_width;
          index += availible_width;
        }

        new_line = String::from(yap[index..].to_string());
        coldex = new_line.len();

        if new_line.len() > 0 {
          new_line.push_str(" ");
          coldex += 1;
        }
      }
    }

    lines.push(new_line.clone().trim_end().to_string());

    self.cached_length = self.body.len() as u16;
    self.cached_width = width;
    self.cached_lines = lines;
  }

  // this is the one you call
  fn as_lines(&mut self, width: u16) -> &Vec<String> {
    // criteria for refreshing the cache
    if width != self.cached_width || self.body.len() as u16 != self.cached_length {
      self.update_cache(width);
    }

    return &self.cached_lines;
  }
}

#[derive(Debug, Default, Clone)]
pub struct Message {
  body: MulitLineString,
  sender: String,
}

#[derive(Default, Debug)]
pub struct Location {
  index: u64,
  offset: i16,
}

#[derive(Debug, Default)]
pub struct Chat {
  // id: u8,
  messages: Vec<Message>,
  location: Location,
  text_input: TextInput,
}

#[derive(Debug, Default)]
pub struct TextInput {
  body: MulitLineString,
  cursor_index: usize,
}

impl TextInput {
  fn render(&mut self, area: Rect, buf: &mut Buffer, _logger: &mut Logger) {
    let block = Block::bordered().border_set(border::THICK);

    // shitty temp padding for the border
    // let mut area = area;
    // area.x += 1;
    // area.width -= 2;
    // area.height -= 2;
    // area.y += 1;

    let vec_lines = self.body.as_lines(area.width - 2).to_vec();
    // logger.log(format!("this is the first line: {}", self.cursor_index));
    let mut lines: Vec<Line> = Vec::new();
    for yap in vec_lines {
      lines.push(Line::from(yap));
    }

    Paragraph::new(lines).block(block).render(area, buf);
  }

  fn insert_char(&mut self, char: char, logger: &mut Logger) {
    logger.log(format!("adding this: {}", char));
    logger.log(format!("before add: {}", self.body.body));
    // some disgusting object-oriented blashphemy going on here
    self.body.body.insert(self.cursor_index, char);
    self.cursor_index += 1;
    logger.log(format!("after add: {}", self.body.body));
  }
  fn delete_char(&mut self) {
    if self.cursor_index == 0 {
      return;
    }

    self.cursor_index -= 1;
    self.body.body.remove(self.cursor_index);
  }
}

pub struct Settings {
  borders: bool,
  message_width_ratio: f32,
  identity: String,
}

impl Settings {
  fn init() -> Self {
    Self {
      borders: true,
      message_width_ratio: 0.8,
      identity: "me".to_string(),
    }
  }
}

pub struct Logger {
  file: File,
}

impl Logger {
  fn init(file_name: &str) -> Self {
    let mut file = OpenOptions::new()
      .write(true)
      .truncate(true)
      .create(true)
      .open(file_name)
      .expect("am i goated?");

    let result = writeln!(file, "=== START OF LOG === ");

    result.expect("am i goated?");

    file = OpenOptions::new()
      .append(true)
      .create(true)
      .open(file_name)
      .expect("kaboom");

    let logger = Logger { file: file };

    logger
  }

  fn log(&mut self, log: String) {
    // no clue why this works
    writeln!(self.file, "{}", log).expect("kaboom");
  }
}

impl Model {
  fn init() -> Self {
    let messages = vec![
      Message {
        body: MulitLineString::init(
          "first message lets make this message super looong jjafkldjaflk it was not long enough last time time to yap fr",
        ),
        sender: String::from("not me"),
      },
      Message {
        body: MulitLineString::init("second message"),
        sender: String::from("me"),
      },
    ];

    let mut chat = Chat::default();

    for message in messages {
      chat.messages.push(message);
    }

    chat.text_input = TextInput::default();

    let chats: Vec<Chat> = vec![chat];

    let mut model = Model::default();
    model.chats = chats;
    model.chat_index = 0;
    model
  }

  // not really needed but it staves off the need for explicit liiftimes a little longer
  fn current_chat(&mut self) -> &mut Chat {
    &mut self.chats[self.chat_index]
  }
}

impl Message {
  fn render(&mut self, area: Rect, buf: &mut Buffer, settings: &Settings, _logger: &mut Logger) {
    let block = Block::bordered().border_set(border::THICK);

    // this ugly shadow cost me a good 15 mins of my life ... but im not changing it
    let mut my_area = area.clone();
    my_area.width = (area.width as f32 * settings.message_width_ratio + 0.5) as u16;
    // let message_width: u16 = (area.width as f32 * settings.message_width_ratio + 0.5) as u16 - 2;

    // let mut new_area = area.clone();
    // new_area.width = area.width;
    //
    // let mut lines: Vec<String> = Vec::new();
    //
    // let mut index = 0;

    // let result = &self.lines;

    let vec_lines: Vec<String> = self.body.as_lines(my_area.width - 2).to_vec();

    // match result {
    //   Some(x) => vec_lines = x.to_vec(),
    //   None => {
    //     panic!("AAAaaaHHHhhh!!!")
    //   }
    // }

    // shrink the message to fit if it does not need mutliple lines
    if vec_lines.len() == 1 {
      my_area.width = vec_lines[0].len() as u16 + 2;
    }

    // "allign" the chat to the right if it was sent by you
    // TODO: should add setting to toggle this behavior
    if settings.identity == self.sender {
      my_area.x += area.width - my_area.width;
    }

    let mut lines: Vec<Line> = Vec::new();
    for yap in vec_lines {
      lines.push(Line::from(yap));
    }

    Paragraph::new(lines).block(block).render(my_area, buf)
    // .wrap(Wrap { trim: true })
  }
}

// impl Message {
// me after i find out this already exists in ratatui: -_-
//
// fn split_into_lines(&mut self, term_width: u16) -> Vec<String> {
//   let mut lines: Vec<String> = Vec::new();
//   let mut new_line = String::from("");
//
//   // collumn index
//   let mut coldex = 0;
//   // let availible_width = (term_width as f32 * settings.message_width_ratio + 0.5) as usize;
//   let availible_width = term_width as usize;
//
//   for yap in self.body.split(" ") {
//     let mut length = yap.len();
//
//     if coldex + yap.len() <= availible_width {
//       new_line.push_str(yap);
//       new_line.push_str(" ");
//       coldex += yap.len() + 1;
//     } else {
//       lines.push(new_line.clone().trim_end().to_string());
//
//       let mut index = 0;
//
//       while length >= availible_width {
//         lines.push(yap[index..index + availible_width].to_string());
//         length -= availible_width;
//         index += availible_width;
//       }
//
//       new_line = String::from(yap[index..].to_string());
//       coldex = new_line.len();
//
//       if new_line.len() > 0 {
//         new_line.push_str(" ");
//         coldex += 1;
//       }
//     }
//   }
//
//   lines.push(new_line.clone().trim_end().to_string());
//
//   lines
// }
// }

fn format_vec(vec: &Vec<String>) -> String {
  let mut output = String::from("[");

  for thing in vec {
    output.push_str(thing);
    output.push_str(", ");
  }

  output.push_str("]");

  return output;
}

impl Chat {
  fn render(&mut self, area: Rect, buf: &mut Buffer, settings: &Settings, logger: &mut Logger) {
    let input_lines = self.text_input.body.as_lines(area.width - 2).len() as u16;
    logger.log("this is our input: ".to_string());
    logger.log(format_vec(self.text_input.body.as_lines(area.width - 2)));

    let layout = Layout::vertical([Constraint::Min(6), Constraint::Length(input_lines + 2)]).split(area);

    // kind of a sketchy shadow here but the layout[1] is used like once
    let area = layout[0];

    let block = Block::bordered().border_set(border::THICK);
    // .title(title.centered())
    // .title_bottom(instructions.centered())
    block.render(area, buf);

    // shitty temp padding for the border
    let mut area = area;
    area.x += 1;
    area.width -= 2;
    area.height -= 2;
    area.y += 1;
    // end shitty tmp padding

    let message_width: u16 = (area.width as f32 * settings.message_width_ratio + 0.5) as u16 - 2;

    let mut index = 0;
    let mut y = self.location.offset * -1;

    while y < area.height as i16 && index < self.messages.len() {
      let message = &mut self.messages[index];

      // only here for testing; remove later; breaks the fun "cache" system
      // message.lines = message.as_lines(message_width);

      // let result = message.body.as_lines(mes);

      // match result {
      //   Some(_x) => {}
      //   None => {
      //     message.lines = Some(message.split_into_lines(message_width));
      //     let this_better_work = &message.lines;
      //     match this_better_work {
      //       Some(_x) => {}
      //       None => panic!("AAAaaaHHHhhh!!!"),
      //     }
      //   }
      // }

      let height = message.body.as_lines(message_width).len() as i16 + 2;

      // let height = min(y + requested_height, area.height);
      let new_area = Rect::new(area.x, area.y + y as u16, area.width, height as u16);

      message.render(new_area, buf, settings, logger);

      index += 1;
      y += height;
    }

    self.text_input.render(layout[1], buf, logger);
  }
}

fn main() -> color_eyre::Result<()> {
  // tui::install_panic_hook();
  let mut terminal = ratatui::init();
  let mut model = Model::init();
  let settings = &Settings::init();

  // regular lumber jack
  let logger = &mut Logger::init("log.txt");
  logger.log("testing".to_string());

  while model.running_state != RunningState::OhShit {
    // Render the current view
    terminal.draw(|f| view(&mut model, f, settings, logger))?;

    // Handle events and map to a Message
    let mut current_msg = handle_event(&model)?;

    // Process updates as long as they return a non-None message
    while current_msg.is_some() {
      current_msg = update(&mut model, current_msg.unwrap(), logger);
    }
  }

  ratatui::restore();
  Ok(())
}

// TODO: gotta figure out how to model chat state

fn view(model: &mut Model, frame: &mut Frame, settings: &Settings, logger: &mut Logger) {
  let title = Line::from(" Counter App Tutorial ".bold());
  let instructions = Line::from(vec![
    " Decrement ".into(),
    "<Left>".blue().bold(),
    " Increment ".into(),
    "<Right>".blue().bold(),
    " Quit ".into(),
    "<Q> ".blue().bold(),
  ]);
  let _block = Block::bordered()
    .title(title.centered())
    .title_bottom(instructions.centered())
    .border_set(border::THICK);

  // let _counter_text = Text::from(vec![Line::from(vec![
  //   "Value: ".into(),
  //   model.counter.to_string().yellow(),
  // ])]);

  let layout = Layout::new(
    Direction::Horizontal,
    vec![Constraint::Percentage(20), Constraint::Percentage(80)],
  )
  .split(frame.area());

  model
    .current_chat()
    .render(layout[1], frame.buffer_mut(), settings, logger);

  //
  // frame.render_widget(
  //   Paragraph::new(message_text).right_aligned().block(block),
  //   frame.area(),
  // );

  // let p = Paragraph::new("A very long text that might not fit the container...")
  //   .wrap(Wrap { trim: true });
  //
  // let test_rect = Rect::new(10, 10, 7, 20);
  // frame.render_widget(p, test_rect);
}

/// Convert Event to Action
///
/// We don't need to pass in a `model` to this function in this example
/// but you might need it as your project evolves
///
/// (the project evolved (pokemon core))
fn handle_event(model: &Model) -> color_eyre::Result<Option<Action>> {
  if event::poll(Duration::from_millis(250))? {
    if let Event::Key(key) = event::read()? {
      if key.kind == event::KeyEventKind::Press {
        return Ok(handle_key(key, model));
      }
    }
  }
  Ok(None)
}

fn handle_key(key: event::KeyEvent, model: &Model) -> Option<Action> {
  match model.mode {
    Mode::Insert => match key.code {
      KeyCode::Esc => Some(Action::SetMode(Mode::Normal)),
      KeyCode::Char(char) => Some(Action::Type(char)),
      // this will not get confusing trust
      KeyCode::Backspace => Some(Action::Backspace),
      _ => None,
    },
    Mode::Normal => match key.code {
      KeyCode::Char('i') => Some(Action::SetMode(Mode::Insert)),
      // KeyCode::Char('k') => Some(Action::Decrement),
      KeyCode::Char('q') => Some(Action::Quit),
      _ => None,
    },
  }
}

fn update(model: &mut Model, msg: Action, logger: &mut Logger) -> Option<Action> {
  match msg {
    // Action::Increment => {
    //   model.counter += 1;
    //   if model.counter > 50 {
    //     return Some(Action::Reset);
    //   }
    // }
    // Action::Decrement => {
    //   model.counter -= 1;
    //   if model.counter < -50 {
    //     return Some(Action::Reset);
    //   }
    // }
    Action::Type(char) => {
      model.current_chat().text_input.insert_char(char, logger);
    }
    Action::Backspace => model.current_chat().text_input.delete_char(),

    Action::SetMode(new_mode) => model.mode = new_mode,

    Action::Quit => {
      // You can handle cleanup and exit here
      // - im ok thanks tho
      model.running_state = RunningState::OhShit;
    }
  };
  None
}

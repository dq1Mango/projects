mod tests;
use core::panic;
use std::fmt::format;
use std::fs::File;
use std::io::Write;
use std::{cmp::min, io::LineWriter, time::Duration, vec};

use ratatui::{
  Frame,
  buffer::Buffer,
  crossterm::event::{self, Event, KeyCode},
  layout::Rect,
  style::Stylize,
  symbols::border,
  text::{Line, Text},
  widgets::{Block, Paragraph, Widget, Wrap},
};
use std::fs::OpenOptions;

#[derive(Debug, Default)]
struct Model {
  counter: i32,
  running_state: RunningState,
  content: Chat,
}

#[derive(Debug, Default, PartialEq, Eq)]
enum RunningState {
  #[default]
  Running,
  Done,
}

#[derive(PartialEq)]
enum Action {
  Increment,
  Decrement,
  Reset,
  Quit,
}

pub struct Settings {
  borders: bool,
  message_width_ratio: f32,
}

#[derive(Debug, Default, Clone)]
pub struct Message {
  body: String,
  lines: Option<Vec<String>>,
  sender: String,
}

#[derive(Default, Debug)]
pub struct Location {
  index: u64,
  offset: i16,
}

#[derive(Debug, Default)]
pub struct Chat {
  id: u8,
  messages: Vec<Message>,
  location: Location,
}

pub struct Logger {
  file: File,
}

impl Logger {
  fn init(file_name: &str) -> Logger {
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
    writeln!(self.file, "{}", log).expect("kaboom");
  }
}

impl Model {
  fn init() -> Model {
    let messages = vec![
      Message {
        body: String::from(
          "first message lets make this message super loooong jfjlkdsjafkldjaflk it was not long enough last time time to yap fr",
        ),
        sender: String::from(""),
        lines: None,
      },
      Message {
        body: String::from("second message"),
        sender: String::from(""),
        lines: None,
      },
    ];

    let mut chat = Chat::default();

    for message in messages {
      chat.messages.push(message);
    }

    let mut app = Model::default();
    app.content = chat;
    app
  }
}

impl Message {
  fn render(&self, area: Rect, buf: &mut Buffer, settings: &Settings, logger: &mut Logger) {
    let block = Block::bordered().border_set(border::THICK);

    let mut my_area = area.clone();
    my_area.width = (area.width as f32 * settings.message_width_ratio) as u16;

    // let mut new_area = area.clone();
    // new_area.width = area.width;
    //
    // let mut lines: Vec<String> = Vec::new();
    //
    // let mut index = 0;

    let result = &self.lines;
    let vec_lines: Vec<String>;

    match result {
      Some(x) => vec_lines = x.to_vec(),
      None => {
        panic!("AAAaaaHHHhhh!!!")
      }
    }

    let mut lines: Vec<Line> = Vec::new();
    for yap in vec_lines {
      lines.push(Line::from(yap));
    }

    Paragraph::new(lines).block(block).render(my_area, buf)
    // .wrap(Wrap { trim: true })
  }
}

impl Message {
  fn split_into_lines(&mut self, term_width: u16) -> Vec<String> {
    let mut lines: Vec<String> = Vec::new();
    let mut new_line = String::from("");

    // collumn index
    let mut coldex = 0;
    // let availible_width = (term_width as f32 * settings.message_width_ratio + 0.5) as usize;
    let availible_width = term_width as usize;

    for yap in self.body.split(" ") {
      let mut length = yap.len();

      if coldex + yap.len() <= availible_width {
        new_line.push_str(yap);
        new_line.push_str(" ");
        coldex += yap.len() + 1;
      } else {
        lines.push(new_line.clone().trim_end().to_string());

        let mut index = 0;

        while length >= availible_width {
          lines.push(yap[index..index + availible_width].to_string());
          length -= availible_width;
          index += availible_width;
        }

        new_line = String::from(yap[index..].to_string());
        coldex = new_line.len() + 1;
        new_line.push_str(" ");
      }
    }

    lines.push(new_line.clone().trim_end().to_string());

    lines
  }
}

impl Chat {
  fn render(&mut self, area: Rect, buf: &mut Buffer, settings: &Settings, logger: &mut Logger) {
    let message_width: u16 = (area.width as f32 * settings.message_width_ratio) as u16 - 2;
    logger.log(format!("width: {}", message_width));

    let mut index = 0;
    let mut y = self.location.offset * -1;

    while y < area.height as i16 && index < self.messages.len() {
      let message = &mut self.messages[index];

      // only here for testing; remove later; breaks the fun "cache" system
      message.lines = Some(message.split_into_lines(message_width));

      let result = &message.lines;

      match result {
        Some(_x) => {}
        None => {
          message.lines = Some(message.split_into_lines(message_width));
          let this_better_work = &message.lines;
          match this_better_work {
            Some(_x) => {}
            None => panic!("AAAaaaHHHhhh!!!"),
          }
        }
      }

      let height = message.lines.as_ref().unwrap().len() as i16 + 2;

      // let height = min(y + requested_height, area.height);
      let new_area = Rect::new(area.x, area.y + y as u16, area.width, height as u16);

      message.render(new_area, buf, settings, logger);

      index += 1;
      y += height;
    }
  }
}

fn main() -> color_eyre::Result<()> {
  // tui::install_panic_hook();
  let mut terminal = ratatui::init();
  let mut model = Model::init();
  let logger = &mut Logger::init("log.txt");
  logger.log("testing".to_string());
  let settings = &Settings {
    borders: true,
    message_width_ratio: 0.8,
  };

  while model.running_state != RunningState::Done {
    // Render the current view
    terminal.draw(|f| view(&mut model, f, settings, logger))?;

    // Handle events and map to a Message
    let mut current_msg = handle_event(&model)?;

    // Process updates as long as they return a non-None message
    while current_msg.is_some() {
      current_msg = update(&mut model, current_msg.unwrap());
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
  let block = Block::bordered()
    .title(title.centered())
    .title_bottom(instructions.centered())
    .border_set(border::THICK);

  let _counter_text = Text::from(vec![Line::from(vec![
    "Value: ".into(),
    model.counter.to_string().yellow(),
  ])]);

  model
    .content
    .render(frame.area(), frame.buffer_mut(), settings, logger);

  // let mut messages: Vec<ratatui::prelude::Line> = Vec::new();
  //
  // for message in model.content.messages.clone() {
  //   messages.push(Line::from(message.body));
  // }
  //
  // let message_text = Text::from(messages);
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
fn handle_event(_: &Model) -> color_eyre::Result<Option<Action>> {
  if event::poll(Duration::from_millis(250))? {
    if let Event::Key(key) = event::read()? {
      if key.kind == event::KeyEventKind::Press {
        return Ok(handle_key(key));
      }
    }
  }
  Ok(None)
}

fn handle_key(key: event::KeyEvent) -> Option<Action> {
  match key.code {
    KeyCode::Char('j') => Some(Action::Increment),
    KeyCode::Char('k') => Some(Action::Decrement),
    KeyCode::Char('q') => Some(Action::Quit),
    _ => None,
  }
}

fn update(model: &mut Model, msg: Action) -> Option<Action> {
  match msg {
    Action::Increment => {
      model.counter += 1;
      if model.counter > 50 {
        return Some(Action::Reset);
      }
    }
    Action::Decrement => {
      model.counter -= 1;
      if model.counter < -50 {
        return Some(Action::Reset);
      }
    }
    Action::Reset => model.counter = 0,
    Action::Quit => {
      // You can handle cleanup and exit here
      model.running_state = RunningState::Done;
    }
  };
  None
}

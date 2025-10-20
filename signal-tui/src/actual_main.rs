#[cfg(test)]
mod tests;

use core::panic;
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
  counter: i32,
  running_state: RunningState,
  content: Chat,
}

#[derive(Debug, Default, PartialEq, Eq)]
enum RunningState {
  #[default]
  Running,
  OhShit,
}

#[derive(PartialEq)]
enum Action {
  Increment,
  Decrement,
  Reset,
  Quit,
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
  // id: u8,
  messages: Vec<Message>,
  location: Location,
}

pub struct Settings {
  borders: bool,
  message_width_ratio: f32,
  identity: String,
}

impl Settings {
  fn init() -> Self {
    Settings {
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
  fn init() -> Model {
    let messages = vec![
      Message {
        body: String::from(
          "first message lets make this message super looong jjafkldjaflk it was not long enough last time time to yap fr",
        ),
        sender: String::from("not me"),
        lines: None,
      },
      Message {
        body: String::from("second message"),
        sender: String::from("me"),
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
  fn render(&self, area: Rect, buf: &mut Buffer, settings: &Settings, _logger: &mut Logger) {
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

impl Message {
  // me after i find out this already exists in ratatui: -_-
  //
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
        coldex = new_line.len();

        if new_line.len() > 0 {
          new_line.push_str(" ");
          coldex += 1;
        }
      }
    }

    lines.push(new_line.clone().trim_end().to_string());

    lines
  }
}

impl Chat {
  fn render(&mut self, area: Rect, buf: &mut Buffer, settings: &Settings, logger: &mut Logger) {
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

  model.content.render(layout[1], frame.buffer_mut(), settings, logger);

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
      model.running_state = RunningState::OhShit;
    }
  };
  None
}
